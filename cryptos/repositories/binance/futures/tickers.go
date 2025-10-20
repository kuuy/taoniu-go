package futures

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "net"
  "net/http"
  "os"
  "sort"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
)

type TickersRepository struct {
  Rdb *redis.Client
  Ctx context.Context
}

type TickerInfo struct {
  Symbol    string  `json:"symbol"`
  Open      float64 `json:"openPrice,string"`
  Price     float64 `json:"lastPrice,string"`
  Change    float64 `json:"-"`
  High      float64 `json:"highPrice,string"`
  Low       float64 `json:"lowPrice,string"`
  Volume    float64 `json:"volume,string"`
  Quota     float64 `json:"quoteVolume,string"`
  CloseTime int64   `json:"closeTime"`
}

func (r *TickersRepository) Flush() error {
  tickers, err := r.Request()
  if err != nil {
    return err
  }
  timestamp := time.Now().UnixMilli()
  pipe := r.Rdb.Pipeline()
  for _, ticker := range tickers {
    redisKey := fmt.Sprintf(config.REDIS_KEY_TICKERS, ticker.Symbol)
    value, err := r.Rdb.HGet(r.Ctx, redisKey, "lasttime").Result()
    if err == nil {
      lasttime, _ := strconv.ParseInt(value, 10, 64)
      if lasttime > ticker.CloseTime {
        continue
      }
    }
    pipe.HMSet(
      r.Ctx,
      redisKey,
      map[string]interface{}{
        "symbol":    ticker.Symbol,
        "open":      ticker.Open,
        "price":     ticker.Price,
        "high":      ticker.High,
        "low":       ticker.Low,
        "volume":    ticker.Volume,
        "quota":     ticker.Quota,
        "change":    ticker.Change,
        "lasttime":  ticker.CloseTime,
        "timestamp": timestamp,
      },
    )
    pipe.ZRem(r.Ctx, "binance:spot:tickers:flush", ticker.Symbol)
  }
  pipe.Exec(r.Ctx)
  return nil
}

func (r *TickersRepository) Request() ([]*TickerInfo, error) {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  proxy := common.GetEnvString("BINANCE_PROXY")
  if proxy != "" {
    tr.DialContext = (&common.ProxySession{
      Proxy: fmt.Sprintf("%v?timeout=5s", proxy),
    }).DialContext
  } else {
    tr.DialContext = (&net.Dialer{}).DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   5 * time.Second,
  }

  url := fmt.Sprintf("%s/fapi/v1/ticker/24hr", os.Getenv("BINANCE_FUTURES_API_ENDPOINT"))
  req, _ := http.NewRequest("GET", url, nil)
  q := req.URL.Query()
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

  var result []*TickerInfo
  json.NewDecoder(resp.Body).Decode(&result)

  for _, ticker := range result {
    if ticker.Open > 0 {
      ticker.Change, _ = decimal.NewFromFloat(ticker.Price).Sub(decimal.NewFromFloat(ticker.Open)).Div(decimal.NewFromFloat(ticker.Open)).Round(4).Float64()
    }
  }

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
		local key = 'binance:futures:realtime:' .. KEYS[i]
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

func (r *TickersRepository) Ranking(
  symbols []string,
  fields []string,
  sortField string,
  sortType int,
  current int,
  pageSize int,
) *RankingResult {
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
		local key = 'binance:futures:realtime:' .. KEYS[i]
		if redis.call('EXISTS', key) == 0 then
			data[i] = false
		else
			data[i] = hmget(key)
		end
	end
	return data
  `)

  sortIdx := -1

  var args []interface{}
  for i, field := range fields {
    if field == sortField {
      sortIdx = i
    }
    args = append(args, field)
  }

  ranking := &RankingResult{}

  if sortIdx == -1 {
    return ranking
  }

  result, _ := script.Run(r.Ctx, r.Rdb, symbols, args...).Result()

  var scores []*RankingScore
  for i := 0; i < len(symbols); i++ {
    item := result.([]interface{})[i]
    if item == nil {
      continue
    }
    if item.([]interface{})[sortIdx] == nil {
      continue
    }
    data := make([]string, len(fields))
    for j := 0; j < len(fields); j++ {
      if item.([]interface{})[j] == nil {
        continue
      }
      data[j] = fmt.Sprintf("%v", item.([]interface{})[j])
    }
    score, _ := strconv.ParseFloat(
      fmt.Sprintf("%v", item.([]interface{})[sortIdx]),
      16,
    )
    scores = append(scores, &RankingScore{
      symbols[i],
      score,
      data,
    })
  }

  if len(scores) == 0 {
    return ranking
  }

  sort.SliceStable(scores, func(i, j int) bool {
    if sortType == -1 {
      return scores[i].Value > scores[j].Value
    } else if sortType == 1 {
      return scores[i].Value < scores[j].Value
    }
    return true
  })

  offset := (current - 1) * pageSize
  endPos := offset + pageSize
  if endPos > len(scores) {
    endPos = len(scores)
  }

  ranking.Total = len(scores)
  for _, score := range scores[offset:endPos] {
    ranking.Data = append(ranking.Data, strings.Join(
      append([]string{score.Symbol}, score.Data...),
      ",",
    ))
  }

  return ranking
}
