package spot

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "net"
  "net/http"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/spot"
)

type KlinesRepository struct {
  Db       *gorm.DB
  Rdb      *redis.Client
  Ctx      context.Context
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

func (r *KlinesRepository) Flush(symbol string, interval string, limit int) error {
  klines, err := r.Request(symbol, interval, limit)
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

  if len(klines) > 0 {
    timestamp := time.Now().Unix()
    r.Rdb.ZAdd(
      r.Ctx,
      fmt.Sprintf(
        "binance:spot:klines:flush:%v",
        interval,
      ),
      &redis.Z{
        float64(timestamp),
        symbol,
      },
    )
  }

  return nil
}

func (r *KlinesRepository) Request(symbol string, interval string, limit int) ([][]interface{}, error) {
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
    Timeout:   time.Duration(8) * time.Second,
  }

  url := "https://api.binance.com/api/v3/klines"
  req, _ := http.NewRequest("GET", url, nil)
  q := req.URL.Query()
  q.Add("symbol", symbol)
  q.Add("interval", interval)
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
  timestamp := time.Now().AddDate(0, 0, -101).Unix()
  r.Db.Where("timestamp < ?", timestamp).Delete(&models.Kline{})
  return nil
}
