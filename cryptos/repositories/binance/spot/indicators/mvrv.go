package indicators

import (
  "bytes"
  "encoding/json"
  "fmt"
  "net/http"
  "strconv"
  "strings"
  "time"

  config "taoniu.local/cryptos/config/binance/spot"
)

type MvrvRepository struct {
  BaseRepository
}

type dashResponse struct {
  Response struct {
    Chart struct {
      Figure struct {
        Data []struct {
          Name string    `json:"name"`
          X    []string  `json:"x"`
          Y    []float64 `json:"y"`
        } `json:"data"`
      } `json:"figure"`
    } `json:"chart"`
  } `json:"response"`
}

func (r *MvrvRepository) fetchRealizedPrice() (float64, error) {
  body := `{"output":"chart.figure","outputs":{"id":"chart","property":"figure"},"inputs":[{"id":"url","property":"pathname","value":"/charts/realized-price/"},{"id":"display","property":"children","value":"xxl 1850px"}],"changedPropIds":["url.pathname","display.children"]}`

  client := &http.Client{Timeout: 15 * time.Second}
  resp, err := client.Post(
    "https://www.bitcoinmagazinepro.com/django_plotly_dash/app/realized_price/_dash-update-component",
    "application/json",
    bytes.NewBufferString(body),
  )
  if err != nil {
    return 0, err
  }
  defer resp.Body.Close()

  var result dashResponse
  if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
    return 0, err
  }

  for _, trace := range result.Response.Chart.Figure.Data {
    if trace.Name == "Realized Price" && len(trace.Y) > 0 {
      return trace.Y[len(trace.Y)-1], nil
    }
  }
  return 0, fmt.Errorf("realized price trace not found")
}

func (r *MvrvRepository) Get(symbol, interval string) (value, price float64, timestamp int64, err error) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(config.REDIS_KEY_INDICATORS, interval, symbol, day)
  val, err := r.Rdb.HGet(r.Ctx, redisKey, "mvrv").Result()
  if err != nil {
    return
  }
  data := strings.Split(val, ",")
  value, _ = strconv.ParseFloat(data[0], 64)
  price, _ = strconv.ParseFloat(data[1], 64)
  timestamp, _ = strconv.ParseInt(data[2], 10, 64)
  return
}

func (r *MvrvRepository) Flush(symbol string, interval string, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "close")
  if err != nil {
    return
  }

  closes := data[0]
  lastIdx := len(timestamps) - 1

  realizedPrice, err := r.fetchRealizedPrice()
  if err != nil {
    return
  }
  if realizedPrice == 0 {
    err = fmt.Errorf("[%s] realized price is zero", symbol)
    return
  }

  mvrv := closes[lastIdx] / realizedPrice

  day, err := r.Day(timestamps[lastIdx] / 1000)
  if err != nil {
    return
  }

  redisKey := fmt.Sprintf(config.REDIS_KEY_INDICATORS, interval, symbol, day)
  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "mvrv",
    fmt.Sprintf(
      "%s,%s,%d",
      strconv.FormatFloat(mvrv, 'f', -1, 64),
      strconv.FormatFloat(closes[lastIdx], 'f', -1, 64),
      timestamps[lastIdx],
    ),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}
