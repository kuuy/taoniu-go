package spot

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "net"
  "net/http"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"

  "taoniu.local/cryptos/common"
)

type TickersRepository struct {
  Rdb      *redis.Client
  Ctx      context.Context
  UseProxy bool
}

func (r *TickersRepository) Flush(symbols []string) error {
  tickers, err := r.Request(symbols)
  if err != nil {
    return err
  }
  timestamp := time.Now().Unix()
  pipe := r.Rdb.Pipeline()
  for _, ticker := range tickers {
    data := ticker.(map[string]interface{})
    symbol := data["symbol"].(string)
    redisKey := fmt.Sprintf("binance:spot:realtime:%s", symbol)
    value, err := r.Rdb.HGet(r.Ctx, redisKey, "price").Result()
    if err == nil {
      lasttime, _ := strconv.ParseInt(value, 10, 64)
      if lasttime > timestamp {
        continue
      }
    }
    price, _ := strconv.ParseFloat(data["lastPrice"].(string), 64)
    open, _ := strconv.ParseFloat(data["openPrice"].(string), 64)
    high, _ := strconv.ParseFloat(data["highPrice"].(string), 64)
    low, _ := strconv.ParseFloat(data["lowPrice"].(string), 64)
    volume, _ := strconv.ParseFloat(data["volume"].(string), 64)
    quota, _ := strconv.ParseFloat(data["quoteVolume"].(string), 64)
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
      "binance:spot:tickers:flush",
      &redis.Z{
        float64(timestamp),
        symbol,
      },
    )
  }
  pipe.Exec(r.Ctx)

  return nil
}

func (r *TickersRepository) Request(symbols []string) ([]interface{}, error) {
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

  url := "https://api.binance.com/api/v3/ticker/24hr"
  req, _ := http.NewRequest("GET", url, nil)
  q := req.URL.Query()
  val, _ := json.Marshal(symbols)
  q.Add("symbols", string(val))
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

  var result []interface{}
  json.NewDecoder(resp.Body).Decode(&result)
  return result, nil
}

func (r *TickersRepository) Gets(symbols []string, fields []string) []string {
  var script = redis.NewScript(`
	local hmget = function (key)
		local hash = {}
		local data = redis.call('HMGET', key, unpack(ARGV))
		for i = 1, #ARGV do
			hash[i] = data[i]
		end
		return hash
	end
	local data = {}
	for i = 1, #KEYS do
		local key = 'binance:spot:realtime:' .. KEYS[i]
		if redis.call('EXISTS', key) == 0 then
			data[i] = false
		else
			data[i] = hmget(key)
		end
	end
	return data
  `)
  args := make([]interface{}, len(fields))
  for i := 0; i < len(fields); i++ {
    args[i] = fields[i]
  }
  result, _ := script.Run(r.Ctx, r.Rdb, symbols, args...).Result()

  tickers := make([]string, len(symbols))
  for i := 0; i < len(symbols); i++ {
    item := result.([]interface{})[i]
    if item == nil {
      continue
    }
    data := make([]string, len(fields))
    for j := 0; j < len(fields); j++ {
      if item.([]interface{})[j] == nil {
        continue
      }
      data[j] = fmt.Sprintf("%v", item.([]interface{})[j])
    }
    tickers[i] = strings.Join(data, ",")
  }

  return tickers
}
