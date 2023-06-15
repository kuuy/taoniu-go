package futures

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

  "taoniu.local/cryptos/common"
)

type TickersRepository struct {
  Rdb      *redis.Client
  Ctx      context.Context
  UseProxy bool
}

func (r *TickersRepository) Flush(symbol string) error {
  ticker, err := r.Request(symbol)
  if err != nil {
    return err
  }
  timestamp := time.Now().Unix()
  pipe := r.Rdb.Pipeline()
  redisKey := fmt.Sprintf("binance:futures:realtime:%s", symbol)
  value, err := r.Rdb.HGet(r.Ctx, redisKey, "price").Result()
  if err == nil {
    lasttime, _ := strconv.ParseInt(value, 10, 64)
    if lasttime > timestamp {
      return errors.New("ticker have been updated")
    }
  }
  price, _ := strconv.ParseFloat(ticker["lastPrice"].(string), 64)
  open, _ := strconv.ParseFloat(ticker["openPrice"].(string), 64)
  high, _ := strconv.ParseFloat(ticker["highPrice"].(string), 64)
  low, _ := strconv.ParseFloat(ticker["lowPrice"].(string), 64)
  volume, _ := strconv.ParseFloat(ticker["volume"].(string), 64)
  quota, _ := strconv.ParseFloat(ticker["quoteVolume"].(string), 64)
  pipe.HMSet(
    r.Ctx,
    redisKey,
    map[string]interface{}{
      "symbol":    symbol,
      "price":     price,
      "open":      open,
      "high":      high,
      "low":       low,
      "volume":    volume,
      "quota":     quota,
      "timestamp": timestamp,
    },
  )
  pipe.ZAdd(
    r.Ctx,
    "binance:futures:tickers:flush",
    &redis.Z{
      float64(timestamp),
      symbol,
    },
  )
  pipe.Exec(r.Ctx)

  return nil
}

func (r *TickersRepository) Request(symbol string) (map[string]interface{}, error) {
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
    Timeout:   time.Duration(5) * time.Second,
  }

  url := "https://fapi.binance.com/fapi/v1/ticker/24hr"
  req, _ := http.NewRequest("GET", url, nil)
  q := req.URL.Query()
  q.Add("symbol", symbol)
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
  return result, nil
}
