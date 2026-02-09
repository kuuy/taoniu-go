package indicators

import (
  "fmt"
  "math"
  "strconv"
  "strings"
  "time"

  "github.com/markcheno/go-talib"

  config "taoniu.local/cryptos/config/binance/futures"
)

type BBandsRepository struct {
  BaseRepository
}

func (r *BBandsRepository) Get(symbol, interval string) (
  b1,
  b2,
  b3,
  w1,
  w2,
  w3,
  price float64,
  timestamp int64,
  err error,
) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  val, err := r.Rdb.HGet(
    r.Ctx,
    redisKey,
    "bbands",
  ).Result()
  if err != nil {
    return
  }
  data := strings.Split(val, ",")
  b1, _ = strconv.ParseFloat(data[0], 64)
  b2, _ = strconv.ParseFloat(data[1], 64)
  b3, _ = strconv.ParseFloat(data[2], 64)
  w1, _ = strconv.ParseFloat(data[3], 64)
  w2, _ = strconv.ParseFloat(data[4], 64)
  w3, _ = strconv.ParseFloat(data[5], 64)
  price, _ = strconv.ParseFloat(data[6], 64)
  timestamp, _ = strconv.ParseInt(data[7], 10, 64)
  return
}

func (r *BBandsRepository) Flush(symbol string, interval string, period int, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "close")
  if err != nil {
    return
  }

  closes := data[0]
  lastIdx := len(timestamps) - 1

  upper, middle, lower := talib.BBands(closes, period, 2, 2, 0)

  b1 := (closes[lastIdx-2] - lower[lastIdx-2]) / math.Max(upper[lastIdx-2]-lower[lastIdx-2], 1e-9)
  b2 := (closes[lastIdx-1] - lower[lastIdx-1]) / math.Max(upper[lastIdx-1]-lower[lastIdx-1], 1e-9)
  b3 := (closes[lastIdx] - lower[lastIdx]) / math.Max(upper[lastIdx]-lower[lastIdx], 1e-9)
  w1 := (upper[lastIdx-2] - lower[lastIdx-2]) / math.Max(middle[lastIdx-2], 1e-9)
  w2 := (upper[lastIdx-1] - lower[lastIdx-1]) / math.Max(middle[lastIdx-1], 1e-9)
  w3 := (upper[lastIdx] - lower[lastIdx]) / math.Max(middle[lastIdx], 1e-9)

  day, err := r.Day(timestamps[lastIdx] / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  r.Rdb.HSet(
    r.Ctx,
    redisKey,
    "bbands",
    fmt.Sprintf(
      "%s,%s,%s,%s,%s,%s,%s,%d",
      strconv.FormatFloat(b1, 'f', -1, 64),
      strconv.FormatFloat(b2, 'f', -1, 64),
      strconv.FormatFloat(b3, 'f', -1, 64),
      strconv.FormatFloat(w1, 'f', -1, 64),
      strconv.FormatFloat(w2, 'f', -1, 64),
      strconv.FormatFloat(w3, 'f', -1, 64),
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
