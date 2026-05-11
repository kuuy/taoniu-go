package futures

import (
  "context"
  "encoding/json"
  "fmt"
  "net"
  "net/http"
  "os"
  "time"

  "github.com/go-redis/redis/v8"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
)

type FundingRateInfo struct {
  Symbol      string  `json:"symbol"`
  FundingRate float64 `json:"lastFundingRate,string"`
  Time        int64   `json:"time"`
}

type FundingRateRepository struct {
  Rdb *redis.Client
  Ctx context.Context
}

func (r *FundingRateRepository) Get(symbol string) (float64, error) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(config.REDIS_KEY_FUNDING_RATE, day)
  val, err := r.Rdb.HGet(r.Ctx, redisKey, symbol).Result()
  if err != nil {
    return 0, err
  }
  var rate float64
  if err = json.Unmarshal([]byte(val), &rate); err != nil {
    return 0, err
  }
  return rate, nil
}

func (r *FundingRateRepository) Flush() (err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  proxy := common.GetEnvString("BINANCE_PROXY")
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

  url := fmt.Sprintf("%s/fapi/v1/premiumIndex", os.Getenv("BINANCE_FUTURES_API_ENDPOINT"))
  req, _ := http.NewRequest("GET", url, nil)
  resp, err := httpClient.Do(req)
  if err != nil {
    return
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    err = fmt.Errorf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode)
    return
  }

  var response []FundingRateInfo
  json.NewDecoder(resp.Body).Decode(&response)

  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(config.REDIS_KEY_FUNDING_RATE, day)

  pipe := r.Rdb.Pipeline()
  for _, item := range response {
    pipe.HSet(r.Ctx, redisKey, item.Symbol, item.FundingRate)
  }
  _, err = pipe.Exec(r.Ctx)
  if err != nil {
    return
  }
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}
