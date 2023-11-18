package dydx

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"
  "net"
  "net/http"
  "os"
  "sort"
  "strconv"
  "strings"
  "taoniu.local/cryptos/common"
  "time"
)

type TickersRepository struct {
  Rdb      *redis.Client
  Ctx      context.Context
  UseProxy bool
}

type TickerInfo struct {
  Symbol string  `json:"market"`
  Price  float64 `json:"close,string"`
  Open   float64 `json:"open,string"`
  High   float64 `json:"high,string"`
  Low    float64 `json:"low,string"`
  Volume float64 `json:"baseVolume,string"`
  Quota  float64 `json:"quoteVolume,string"`
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
		local key = 'dydx:realtime:' .. KEYS[i]
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
		local key = 'dydx:realtime:' .. KEYS[i]
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

func (r *TickersRepository) Flush() error {
  tickers, err := r.Request()
  if err != nil {
    return err
  }
  timestamp := time.Now().Unix()
  pipe := r.Rdb.Pipeline()
  for _, ticker := range tickers {
    redisKey := fmt.Sprintf("dydx:realtime:%s", ticker.Symbol)
    change, _ := decimal.NewFromFloat(ticker.Price).Sub(decimal.NewFromFloat(ticker.Open)).Div(decimal.NewFromFloat(ticker.Open)).Round(4).Float64()
    values := map[string]interface{}{
      "symbol":    ticker.Symbol,
      "price":     ticker.Price,
      "open":      ticker.Open,
      "high":      ticker.High,
      "low":       ticker.Low,
      "volume":    ticker.Volume,
      "quota":     ticker.Quota,
      "change":    change,
      "lasttime":  timestamp,
      "timestamp": timestamp,
    }
    pipe.HMSet(
      r.Ctx,
      redisKey,
      values,
    )
  }
  pipe.Exec(r.Ctx)
  return nil
}

func (r *TickersRepository) Request() ([]*TickerInfo, error) {
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

  url := fmt.Sprintf("%s/v3/stats", os.Getenv("DYDX_API_ENDPOINT"))
  req, _ := http.NewRequest("GET", url, nil)
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

  var result map[string]map[string]TickerInfo
  json.NewDecoder(resp.Body).Decode(&result)

  if _, ok := result["markets"]; !ok {
    return nil, errors.New("invalid response")
  }

  var tickers []*TickerInfo
  for _, item := range result["markets"] {
    tickers = append(tickers, &TickerInfo{
      Symbol: item.Symbol,
      Price:  item.Price,
      Open:   item.Open,
      High:   item.High,
      Low:    item.Low,
      Volume: item.Volume,
      Quota:  item.Quota,
    })
  }

  return tickers, nil
}
