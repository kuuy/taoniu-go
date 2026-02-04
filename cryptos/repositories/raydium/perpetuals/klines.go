package perpetuals

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
  models "taoniu.local/cryptos/models/raydium/perpetuals"
)

type KlinesRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

type OrderlyHistoryResponse struct {
  S string    `json:"s"`
  T []int64   `json:"t"`
  O []float64 `json:"o"`
  H []float64 `json:"h"`
  L []float64 `json:"l"`
  C []float64 `json:"c"`
  V []float64 `json:"v"`
}

func (r *KlinesRepository) Flush(
  symbol string,
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

  var resolution string
  switch interval {
  case "1m":
    resolution = "1"
  case "15m":
    resolution = "15"
  case "4h":
    resolution = "240"
  case "1d":
    resolution = "D"
  default:
    resolution = interval
  }

  location := "https://api.orderly.org/v1/tv/history"
  params := url.Values{}
  params.Add("symbol", symbol)
  params.Add("resolution", resolution)
  params.Add("from", fmt.Sprintf("%v", startTime/1000))
  params.Add("to", fmt.Sprintf("%v", endTime/1000))

  fmt.Printf("fetching raydium perpetuals klines from: %s?%s\n", location, params.Encode())

  req, err := http.NewRequestWithContext(r.Ctx, "GET", location, nil)
  if err != nil {
    return err
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

  var response OrderlyHistoryResponse
  if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
    return fmt.Errorf("decode error: %v", err)
  }

  if response.S != "ok" && response.S != "no_data" {
    return fmt.Errorf("failed to fetch klines: api returned s = %s", response.S)
  }

  for i, t := range response.T {
    timestamp := t * 1000
    var entity models.Kline
    err = r.Db.Where("symbol=? AND interval=? AND timestamp=?", symbol, interval, timestamp).Take(&entity).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
      entity = models.Kline{
        ID:        xid.New().String(),
        Symbol:    symbol,
        Interval:  interval,
        Open:      response.O[i],
        High:      response.H[i],
        Low:       response.L[i],
        Close:     response.C[i],
        Volume:    response.V[i],
        Quota:     0, // tv history doesn't provide amount/quota
        Timestamp: timestamp,
      }
      r.Db.Create(&entity)
    } else {
      r.Db.Model(&entity).Updates(map[string]interface{}{
        "open":   response.O[i],
        "high":   response.H[i],
        "low":    response.L[i],
        "close":  response.C[i],
        "volume": response.V[i],
      })
    }
  }

  return nil
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
