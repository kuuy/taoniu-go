package indicators

import (
  "fmt"
  "strconv"
  config "taoniu.local/cryptos/config/binance/futures"
  "time"
)

type PivotRepository struct {
  BaseRepository
}

func (r *PivotRepository) Get(symbol, interval string) (
  r3 float64,
  r2 float64,
  r1 float64,
  s1 float64,
  s2 float64,
  s3 float64,
  err error,
) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )

  fields := []string{
    "r3",
    "r2",
    "r1",
    "s1",
    "s2",
    "s3",
  }
  data, err := r.Rdb.HMGet(
    r.Ctx,
    redisKey,
    fields...,
  ).Result()
  if err != nil {
    return
  }

  for i := 0; i < len(fields); i++ {
    switch fields[i] {
    case "r3":
      r3, _ = strconv.ParseFloat(data[i].(string), 64)
    case "r2":
      r2, _ = strconv.ParseFloat(data[i].(string), 64)
    case "r1":
      r1, _ = strconv.ParseFloat(data[i].(string), 64)
    case "s1":
      s1, _ = strconv.ParseFloat(data[i].(string), 64)
    case "s2":
      s2, _ = strconv.ParseFloat(data[i].(string), 64)
    case "s3":
      s3, _ = strconv.ParseFloat(data[i].(string), 64)
    }
  }

  return
}

func (r *PivotRepository) Flush(symbol string, interval string) (err error) {
  data, timestamp, err := r.Kline(symbol, interval, "close", "high", "low")
  if err != nil {
    return
  }

  price, high, low := data[0], data[1], data[2]

  s1 := price*2 - high
  r1 := price*2 - low
  s2 := price - (r1 - s1)
  r2 := price + (r1 - s1)
  s3 := low - 2*(high-price)
  r3 := high + 2*(price-low)

  day, err := r.Day(timestamp / 1000)
  if err != nil {
    return err
  }

  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  r.Rdb.HMSet(
    r.Ctx,
    redisKey,
    map[string]interface{}{
      "r3": r3,
      "r2": r2,
      "r1": r1,
      "s1": s1,
      "s2": s2,
      "s3": s3,
    },
  )
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}
