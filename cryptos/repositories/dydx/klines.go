package dydx

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "github.com/nats-io/nats.go"
  "log"
  "net"
  "net/http"
  "os"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/dydx"
  models "taoniu.local/cryptos/models/dydx"
)

type KlinesRepository struct {
  Db       *gorm.DB
  Rdb      *redis.Client
  Ctx      context.Context
  Nats     *nats.Conn
  UseProxy bool
}

type KlineInfo struct {
  Symbol    string
  Interval  string
  Open      float64
  Close     float64
  High      float64
  Low       float64
  Volume    float64
  Quota     float64
  Timestamp int64
}

func (r *KlinesRepository) Series(symbol string, interval string, timestamp int64, limit int) []interface{} {
  var klines []*models.Kline
  r.Db.Where(
    "symbol=? AND interval=? AND timestamp<?",
    symbol,
    interval,
    timestamp,
  ).Order("timestamp desc").Limit(limit).Find(&klines)

  series := make([]interface{}, len(klines))
  for i, kline := range klines {
    series[i] = []interface{}{
      kline.Open,
      kline.High,
      kline.Low,
      kline.Close,
      kline.Timestamp,
    }
  }
  return series
}

func (r *KlinesRepository) Count(symbol string, interval string) int64 {
  var total int64
  r.Db.Model(&models.Kline{}).Where("symbol=? AND interval=?", symbol, interval).Count(&total)
  return total
}

func (r *KlinesRepository) History(
  symbol string,
  interval string,
  from int64,
  to int64,
  limit int,
) []*models.Kline {
  var klines []*models.Kline
  r.Db.Model(&models.Kline{}).Where(
    "symbol=? AND interval=? AND timestamp BETWEEN ? AND ?",
    symbol,
    interval,
    from,
    to,
  ).Order("timestamp desc").Limit(limit).Find(&klines)
  return klines
}

func (r *KlinesRepository) Flush(symbol string, interval string, endtime int64, limit int) error {
  klines, err := r.Request(symbol, interval, endtime, limit)
  if err != nil {
    return err
  }

  for _, kline := range klines {
    var entity models.Kline
    result := r.Db.Where(
      "symbol=? AND interval=? AND timestamp=?",
      symbol,
      interval,
      kline.Timestamp,
    ).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      entity = models.Kline{
        ID:        xid.New().String(),
        Symbol:    symbol,
        Interval:  interval,
        Open:      kline.Open,
        Close:     kline.Close,
        High:      kline.High,
        Low:       kline.Low,
        Volume:    kline.Volume,
        Quota:     kline.Quota,
        Timestamp: kline.Timestamp,
      }
      r.Db.Create(&entity)
    } else {
      entity.Open = kline.Open
      entity.Close = kline.Close
      entity.High = kline.High
      entity.Low = kline.Low
      entity.Volume = kline.Volume
      entity.Quota = kline.Quota
      entity.Timestamp = kline.Timestamp
      r.Db.Model(&models.Kline{ID: entity.ID}).Updates(entity)
    }
  }

  if len(klines) > 0 && len(klines) < limit {
    endtime := klines[len(klines)-1].Timestamp - r.Timestep(interval)
    return r.Flush(symbol, interval, endtime, limit-len(klines))
  }

  message, _ := json.Marshal(map[string]interface{}{
    "symbol":   symbol,
    "interval": interval,
  })
  r.Nats.Publish(config.NATS_KLINES_UPDATE, message)
  r.Nats.Flush()

  return nil
}

func (r *KlinesRepository) Fix(symbol string, interval string, limit int) error {
  var klines []*models.Kline
  r.Db.Select(
    []string{"close", "high", "low", "volume", "timestamp"},
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  )

  if len(klines) == 0 {
    return nil
  }

  timestamp := r.Timestamp(interval)
  timestep := r.Timestep(interval)
  lasttime := klines[0].Timestamp

  if timestamp < lasttime {
    timestamp = lasttime
  }

  var endtime int64
  var count int
  if lasttime != timestamp {
    endtime = timestamp
    count = int((timestamp - lasttime) / timestep)
  }

  for i := 1; i < len(klines); i++ {
    if lasttime-klines[i].Timestamp != timestep {
      if endtime == 0 {
        endtime = lasttime
        count = int((endtime - klines[i].Timestamp) / timestep)
      }
    } else {
      if endtime > 0 {
        err := r.Flush(symbol, interval, endtime, count)
        if err != nil {
          log.Println("klines fix error", err.Error())
        }
        endtime = 0
      }
    }
    count++
    lasttime = klines[i].Timestamp
  }

  if count > limit {
    count = limit
  }

  if endtime > 0 {
    log.Println("klines fix", symbol, interval, endtime, count)
    err := r.Flush(symbol, interval, endtime, count)
    if err != nil {
      log.Println("klines fix error", err.Error())
    }
  } else if limit > count {
    log.Println("klines fix", symbol, interval, lasttime, limit-count)
    err := r.Flush(symbol, interval, lasttime, limit-count)
    if err != nil {
      log.Println("klines fix error", err.Error())
    }
  }

  return nil
}

func (r *KlinesRepository) Request(symbol string, interval string, endtime int64, limit int) ([]*KlineInfo, error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  if r.UseProxy {
    session := &common.ProxySession{
      Proxy: "socks5://127.0.0.1:1088?timeout=5s",
    }
    tr.DialContext = session.DialContext
  } else {
    session := &net.Dialer{}
    tr.DialContext = session.DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(3) * time.Second,
  }

  var resolution string
  if interval == "1m" {
    resolution = "1MIN"
  } else if interval == "4h" {
    resolution = "4HOURS"
  } else {
    resolution = "1DAY"
  }

  url := fmt.Sprintf("%s/v3/candles/%s", os.Getenv("DYDX_API_ENDPOINT"), symbol)
  req, _ := http.NewRequest("GET", url, nil)
  q := req.URL.Query()
  q.Add("resolution", resolution)
  if endtime > 0 {
    q.Add("toISO", time.Unix(0, endtime*int64(time.Millisecond)).UTC().Format("2006-01-02T15:04:05.000Z"))
  }
  if limit < 100 {
    q.Add("limit", fmt.Sprintf("%v", limit))
  }
  req.URL.RawQuery = q.Encode()
  resp, err := httpClient.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return nil, errors.New(
      fmt.Sprintf(
        "request error: status[%s] code[%d]",
        resp.Status,
        resp.StatusCode,
      ),
    )
  }

  var result map[string]interface{}
  json.NewDecoder(resp.Body).Decode(&result)

  if _, ok := result["candles"]; !ok {
    return nil, errors.New("invalid response")
  }

  var klines []*KlineInfo
  for _, kline := range result["candles"].([]interface{}) {
    kline := kline.(map[string]interface{})
    startedAt, _ := time.Parse("2006-01-02T15:04:05.000Z", kline["startedAt"].(string))
    klineInfo := &KlineInfo{}
    klineInfo.Symbol = symbol
    klineInfo.Interval = interval
    klineInfo.Open, _ = strconv.ParseFloat(kline["open"].(string), 64)
    klineInfo.Close, _ = strconv.ParseFloat(kline["close"].(string), 64)
    klineInfo.High, _ = strconv.ParseFloat(kline["high"].(string), 64)
    klineInfo.Low, _ = strconv.ParseFloat(kline["low"].(string), 64)
    klineInfo.Volume, _ = strconv.ParseFloat(kline["usdVolume"].(string), 64)
    klineInfo.Quota, _ = strconv.ParseFloat(kline["baseTokenVolume"].(string), 64)
    klineInfo.Timestamp = startedAt.UTC().UnixMilli()
    klines = append(klines, klineInfo)
  }

  return klines, nil
}

func (r *KlinesRepository) Clean() error {
  var timestamp int64

  timestamp = time.Now().AddDate(0, 0, -1).UnixMilli()
  r.Db.Where("interval = ? AND timestamp < ?", "1m", timestamp).Delete(&models.Kline{})

  timestamp = time.Now().AddDate(0, 0, -7).UnixMilli()
  r.Db.Where("interval = ? AND timestamp < ?", "15m", timestamp).Delete(&models.Kline{})

  timestamp = time.Now().AddDate(0, 0, -21).UnixMilli()
  r.Db.Where("interval = ? AND timestamp < ?", "4h", timestamp).Delete(&models.Kline{})

  timestamp = time.Now().AddDate(0, 0, -100).UnixMilli()
  r.Db.Where("interval = ? AND timestamp < ?", "1d", timestamp).Delete(&models.Kline{})

  return nil
}

func (r *KlinesRepository) Timestep(interval string) int64 {
  if interval == "1m" {
    return 60000
  } else if interval == "15m" {
    return 900000
  } else if interval == "4h" {
    return 14400000
  }
  return 86400000
}

func (r *KlinesRepository) Timestamp(interval string) int64 {
  now := time.Now().UTC()
  duration := -time.Second * time.Duration(now.Second())
  if interval == "1m" {
    duration = duration - time.Minute
  } else if interval == "15m" {
    minute, _ := decimal.NewFromInt(int64(now.Minute())).Div(decimal.NewFromInt(15)).Floor().Mul(decimal.NewFromInt(15)).Float64()
    duration = duration - time.Minute*time.Duration(now.Minute()-int(minute))
  } else if interval == "4h" {
    hour, _ := decimal.NewFromInt(int64(now.Hour())).Div(decimal.NewFromInt(4)).Floor().Mul(decimal.NewFromInt(4)).Float64()
    duration = duration - time.Hour*time.Duration(now.Hour()-int(hour)) - time.Minute*time.Duration(now.Minute())
  } else {
    duration = duration - time.Hour*time.Duration(now.Hour()) - time.Minute*time.Duration(now.Minute())
  }
  return now.Add(duration).Unix() * 1000
}
