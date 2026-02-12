package indicators

import (
  "fmt"
  "strconv"
  "strings"
  "time"

  "github.com/markcheno/go-talib"

  config "taoniu.local/cryptos/config/binance/futures"
)

type ZlemaRepository struct {
  BaseRepository
}

func (r *ZlemaRepository) Get(symbol, interval string) (
  prev,
  current,
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
    "zlema",
  ).Result()
  if err != nil {
    return
  }
  data := strings.Split(val, ",")
  prev, _ = strconv.ParseFloat(data[0], 64)
  current, _ = strconv.ParseFloat(data[1], 64)
  price, _ = strconv.ParseFloat(data[2], 64)
  timestamp, _ = strconv.ParseInt(data[3], 10, 64)
  return
}

func (r *ZlemaRepository) Flush(symbol string, interval string, period int, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "close")
  if err != nil {
    return
  }

  closes := data[0]
  lastIdx := len(timestamps) - 1

  lag := (period - 1) / 2

  var zdata []float64
  for i := lag; i <= lastIdx; i++ {
    zdata = append(zdata, closes[i]+(closes[i]-closes[i-lag]))
  }

  result := talib.Ema(zdata, period)

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
    "zlema",
    fmt.Sprintf(
      "%v,%v,%v,%d",
      strconv.FormatFloat(result[len(result)-2], 'f', -1, 64),
      strconv.FormatFloat(result[len(result)-1], 'f', -1, 64),
      strconv.FormatFloat(closes[lastIdx], 'f', -1, 64),
      timestamps[lastIdx],
    ),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}
