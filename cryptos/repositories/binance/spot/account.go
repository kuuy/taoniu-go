package spot

import (
  "context"
  "crypto/hmac"
  "crypto/sha256"
  "encoding/json"
  "errors"
  "fmt"
  "net"
  "net/http"
  "net/url"
  "os"
  "slices"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
)

type AccountRepository struct {
  Db   *gorm.DB
  Rdb  *redis.Client
  Ctx  context.Context
  Nats *nats.Conn
}

func (r *AccountRepository) Flush() (err error) {
  account, err := r.Request()
  if err != nil {
    return
  }
  oldCurrencies, _ := r.Rdb.SMembers(r.Ctx, config.REDIS_KEY_CURRENCIES).Result()
  var currencies []string
  for _, coin := range account.Balances {
    free, _ := strconv.ParseFloat(coin.Free, 64)
    locked, _ := strconv.ParseFloat(coin.Locked, 64)
    if free <= 0.0 {
      r.Rdb.SRem(r.Ctx, config.REDIS_KEY_CURRENCIES, coin.Asset)
      r.Rdb.Del(r.Ctx, fmt.Sprintf(config.REDIS_KEY_BALANCE, coin.Asset))
      continue
    }
    r.Rdb.SAdd(r.Ctx, config.REDIS_KEY_CURRENCIES, coin.Asset)
    r.Rdb.HMSet(
      r.Ctx,
      fmt.Sprintf(config.REDIS_KEY_BALANCE, coin.Asset),
      map[string]interface{}{
        "free":   free,
        "locked": locked,
      },
    )
    currencies = append(currencies, coin.Asset)
  }

  for _, currency := range oldCurrencies {
    if !slices.Contains(currencies, currency) {
      r.Rdb.SRem(r.Ctx, config.REDIS_KEY_CURRENCIES, currency)
      r.Rdb.Del(r.Ctx, fmt.Sprintf(config.REDIS_KEY_BALANCE, currency))
    }
  }

  return
}

func (r *AccountRepository) Balance(asset string) (map[string]float64, error) {
  fields := []string{
    "free",
    "locked",
  }
  data, _ := r.Rdb.HMGet(
    r.Ctx,
    fmt.Sprintf(
      config.REDIS_KEY_BALANCE,
      asset,
    ),
    fields...,
  ).Result()
  balance := map[string]float64{}
  for i, field := range fields {
    if data[i] == nil {
      return nil, errors.New("balance not exists")
    }
    balance[field], _ = strconv.ParseFloat(data[i].(string), 64)
  }
  return balance, nil
}

func (r *AccountRepository) Request() (result *AccountInfo, err error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  proxy := common.GetEnvString("BINANCE_PROXY")
  if proxy != "" {
    tr.DialContext = (&common.ProxySession{
      Proxy: fmt.Sprintf("%v?timeout=3s", proxy),
    }).DialContext
  } else {
    tr.DialContext = (&net.Dialer{}).DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   3 * time.Second,
  }

  params := url.Values{}
  params.Add("recvWindow", "60000")

  timestamp := time.Now().UnixMilli()
  params.Add("timestamp", fmt.Sprintf("%v", timestamp))

  mac := hmac.New(sha256.New, []byte(os.Getenv("BINANCE_SPOT_ACCOUNT_API_SECRET")))
  _, err = mac.Write([]byte(params.Encode()))
  if err != nil {
    return nil, err
  }
  signature := mac.Sum(nil)
  params.Add("signature", fmt.Sprintf("%x", signature))

  url := fmt.Sprintf("%s/api/v3/account", os.Getenv("BINANCE_SPOT_API_ENDPOINT"))
  req, _ := http.NewRequest("GET", url, nil)
  req.URL.RawQuery = params.Encode()
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("X-MBX-APIKEY", os.Getenv("BINANCE_SPOT_ACCOUNT_API_KEY"))
  resp, err := httpClient.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  if resp.StatusCode >= http.StatusBadRequest {
    var apiErr *common.BinanceAPIError
    err = json.NewDecoder(resp.Body).Decode(&apiErr)
    if err == nil {
      err = apiErr
      return
    }
  }

  if resp.StatusCode != http.StatusOK {
    err = errors.New(
      fmt.Sprintf(
        "request error: status[%s] code[%d]",
        resp.Status,
        resp.StatusCode,
      ),
    )
    return
  }

  json.NewDecoder(resp.Body).Decode(&result)
  return
}
