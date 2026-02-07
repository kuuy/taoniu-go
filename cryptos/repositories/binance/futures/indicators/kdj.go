package indicators

import (
  "fmt"
  "strconv"
  "strings"
  "time"

  "github.com/markcheno/go-talib"

  config "taoniu.local/cryptos/config/binance/futures"
)

type KdjRepository struct {
  BaseRepository
}

func (r *KdjRepository) Get(symbol, interval string) (
  slowk,
  slowd,
  slowj,
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
    "kdj",
  ).Result()
  if err != nil {
    return
  }
  data := strings.Split(val, ",")
  slowk, _ = strconv.ParseFloat(data[0], 64)
  slowd, _ = strconv.ParseFloat(data[1], 64)
  slowj, _ = strconv.ParseFloat(data[2], 64)
  price, _ = strconv.ParseFloat(data[3], 64)
  timestamp, _ = strconv.ParseInt(data[4], 10, 64)
  return
}

func (r *KdjRepository) Flush(symbol string, interval string, longPeriod int, shortPeriod int, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "close", "high", "low")
  if err != nil {
    return
  }

  closes := data[0]
  highs := data[1]
  lows := data[2]
  lastIdx := len(timestamps) - 1

  slowk, slowd := talib.Stoch(highs, lows, closes, longPeriod, shortPeriod, 0, shortPeriod, 0)
  slowj := 3*slowk[lastIdx] - 2*slowd[lastIdx]

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
    "kdj",
    fmt.Sprintf(
      "%s,%s,%s,%s,%d",
      strconv.FormatFloat(slowk[lastIdx], 'f', -1, 64),
      strconv.FormatFloat(slowd[lastIdx], 'f', -1, 64),
      strconv.FormatFloat(slowj, 'f', -1, 64),
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
