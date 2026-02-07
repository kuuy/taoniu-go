package indicators

import (
  "fmt"
  "strconv"
  "time"

  "github.com/markcheno/go-talib"

  config "taoniu.local/cryptos/config/binance/futures"
)

type RsiRepository struct {
  BaseRepository
}

func (r *RsiRepository) Get(symbol, interval string) (result float64, err error) {
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
    "rsi",
  ).Result()
  if err != nil {
    return
  }
  result, err = strconv.ParseFloat(val, 64)
  return
}

func (r *RsiRepository) Flush(symbol string, interval string, period int, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "close")
  if err != nil {
    return
  }

  closes := data[0]
  lastIdx := len(timestamps) - 1

  result := talib.Rsi(closes, period)

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
    "rsi",
    strconv.FormatFloat(result[lastIdx], 'f', -1, 64),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}
