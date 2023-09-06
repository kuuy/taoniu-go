package spot

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "log"
  "net"
  "net/http"
  "os"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  models "taoniu.local/cryptos/models/binance/spot"
)

type KlinesRepository struct {
  Db       *gorm.DB
  Rdb      *redis.Client
  Ctx      context.Context
  Nats     *nats.Conn
  UseProxy bool
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

func (r *KlinesRepository) Flush(symbol string, interval string, endtime int64, limit int) error {
  klines, err := r.Request(symbol, interval, endtime, limit)
  if err != nil {
    return err
  }
  for _, kline := range klines {
    open, _ := strconv.ParseFloat(kline[1].(string), 64)
    close, _ := strconv.ParseFloat(kline[4].(string), 64)
    high, _ := strconv.ParseFloat(kline[2].(string), 64)
    low, _ := strconv.ParseFloat(kline[3].(string), 64)
    volume, _ := strconv.ParseFloat(kline[5].(string), 64)
    quota, _ := strconv.ParseFloat(kline[7].(string), 64)
    timestamp := int64(kline[0].(float64))
    var entity models.Kline
    result := r.Db.Where(
      "symbol=? AND interval=? AND timestamp=?",
      symbol,
      interval,
      timestamp,
    ).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      entity = models.Kline{
        ID:        xid.New().String(),
        Symbol:    symbol,
        Interval:  interval,
        Open:      open,
        Close:     close,
        High:      high,
        Low:       low,
        Volume:    volume,
        Quota:     quota,
        Timestamp: timestamp,
      }
      r.Db.Create(&entity)
    } else {
      entity.Open = open
      entity.Close = close
      entity.High = high
      entity.Low = low
      entity.Volume = volume
      entity.Quota = quota
      entity.Timestamp = timestamp
      r.Db.Model(&models.Kline{ID: entity.ID}).Updates(entity)
    }
  }

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

  message, _ := json.Marshal(map[string]interface{}{
    "symbol":   symbol,
    "interval": interval,
  })
  r.Nats.Publish(config.NATS_KLINES_UPDATE, message)
  r.Nats.Flush()

  return nil
}

func (r *KlinesRepository) Request(symbol string, interval string, endtime int64, limit int) ([][]interface{}, error) {
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

  url := fmt.Sprintf("%s/api/v3/klines", os.Getenv("BINANCE_SPOT_API_ENDPOINT"))
  req, _ := http.NewRequest("GET", url, nil)
  q := req.URL.Query()
  q.Add("symbol", symbol)
  q.Add("interval", interval)
  if endtime > 0 {
    q.Add("endTime", fmt.Sprintf("%v", endtime))
  }
  q.Add("limit", fmt.Sprintf("%v", limit))
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

  var result [][]interface{}
  json.NewDecoder(resp.Body).Decode(&result)
  return result, nil
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
  }
  if interval == "4h" {
    return 14400000
  }
  return 86400000
}

func (r *KlinesRepository) Timestamp(interval string) int64 {
  now := time.Now().UTC()
  duration := -time.Second * time.Duration(now.Second())
  if interval == "4h" {
    hour, _ := decimal.NewFromInt(int64(now.Hour())).Div(decimal.NewFromInt(4)).Floor().Mul(decimal.NewFromInt(4)).Float64()
    duration = duration - time.Hour*time.Duration(now.Hour()-int(hour)) - time.Minute*time.Duration(now.Minute())
  } else if interval == "1d" {
    duration = duration - time.Hour*time.Duration(now.Hour()) - time.Minute*time.Duration(now.Minute())
  }
  return now.Add(duration).Unix() * 1000
}
