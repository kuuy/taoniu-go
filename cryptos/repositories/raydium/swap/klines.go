package swap

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "net"
  "net/http"
  "net/url"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/raydium/swap"
)

type KlinesRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *KlinesRepository) Get(symbol string, interval string, timestamp int64) (kline *models.Kline, err error) {
  err = r.Db.Where("symbol=? AND interval=? AND timestamp=?", symbol, interval, timestamp).Take(&kline).Error
  return
}

func (r *KlinesRepository) Flush(
  symbol,
  baseAddress,
  quoteAddress,
  interval string,
  startTime,
  endTime int64,
) error {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  proxy := common.GetEnvString("RAYDIUM_PROXY")
  if proxy != "" {
    tr.DialContext = (&common.ProxySession{
      Proxy: fmt.Sprintf("%v?timeout=30s", proxy),
    }).DialContext
  } else {
    tr.DialContext = (&net.Dialer{}).DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   30 * time.Second,
  }

  headers := map[string]string{
    "X-API-KEY": common.GetEnvString("BIRDEYE_API_KEY"),
    "accept":    "application/json",
    "x-chain":   "solana",
  }

  location := "https://public-api.birdeye.so/defi/ohlcv/base_quote"
  params := url.Values{}
  params.Add("base_address", baseAddress)
  params.Add("quote_address", quoteAddress)
  params.Add("type", interval)
  params.Add("time_from", fmt.Sprintf("%v", startTime/1000))
  params.Add("time_to", fmt.Sprintf("%v", endTime/1000))
  fmt.Printf("fetching klines from: %s\n", location)

  req, err := http.NewRequestWithContext(r.Ctx, "GET", location, nil)
  if err != nil {
    return err
  }
  for key, val := range headers {
    req.Header.Set(key, val)
  }
  req.URL.RawQuery = params.Encode()
  resp, err := httpClient.Do(req)
  if err != nil {
    return err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode)
  }

  var response *KlinesListingsResponse
  if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
    return fmt.Errorf("decode error: %v", err)
  }

  if !response.Success {
    return fmt.Errorf("failed to fetch klines: api returned success = false")
  }

  for _, klineInfo := range response.Data.Items {
    timestamp := klineInfo.Timestamp * 1000
    var entity models.Kline
    err = r.Db.Where("symbol=? AND interval=? AND timestamp=?", symbol, interval, timestamp).Take(&entity).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
      entity = models.Kline{
        ID:        xid.New().String(),
        Symbol:    symbol,
        Interval:  interval,
        Open:      klineInfo.Open,
        High:      klineInfo.High,
        Low:       klineInfo.Low,
        Close:     klineInfo.Close,
        Volume:    klineInfo.Volume,
        Quota:     klineInfo.Quota,
        Timestamp: timestamp,
      }
      r.Db.Create(&entity)
    } else {
      r.Db.Model(&entity).Updates(map[string]interface{}{
        "open":   klineInfo.Open,
        "high":   klineInfo.High,
        "low":    klineInfo.Low,
        "close":  klineInfo.Close,
        "volume": klineInfo.Volume,
        "quota":  klineInfo.Quota,
      })
    }
  }

  return nil
}

func (r *KlinesRepository) Update(kline *models.Kline, column string, value interface{}) (err error) {
  return r.Db.Model(&kline).Update(column, value).Error
}

func (r *KlinesRepository) Updates(kline *models.Kline, values map[string]interface{}) (err error) {
  return r.Db.Model(&kline).Updates(values).Error
}

func (r *KlinesRepository) Timestep(interval string) int64 {
  switch interval {
  case "1m":
    return 60000
  case "15m":
    return 900000
  case "4h":
    return 14400000
  }
  return 86400000
}

func (r *KlinesRepository) Timestamp(interval string) int64 {
  now := time.Now().UTC()
  duration := -time.Second * time.Duration(now.Second())
  switch interval {
  case "15m":
    minute, _ := decimal.NewFromInt(int64(now.Minute())).Div(decimal.NewFromInt(15)).Floor().Mul(decimal.NewFromInt(15)).Float64()
    duration = duration - time.Minute*time.Duration(now.Minute()-int(minute))
  case "4h":
    hour, _ := decimal.NewFromInt(int64(now.Hour())).Div(decimal.NewFromInt(4)).Floor().Mul(decimal.NewFromInt(4)).Float64()
    duration = duration - time.Hour*time.Duration(now.Hour()-int(hour)) - time.Minute*time.Duration(now.Minute())
  case "1d":
    duration = duration - time.Hour*time.Duration(now.Hour()) - time.Minute*time.Duration(now.Minute())
  }
  return now.Add(duration).Unix() * 1000
}
