package indicators

import (
  "fmt"
  "strings"
  "time"

  config "taoniu.local/cryptos/config/binance/futures"
)

type FVGRepository struct {
  BaseRepository
}

func (r *FVGRepository) Get(symbol, interval string) (result float64, err error) {
  return
}

func (r *FVGRepository) Flush(symbol string, interval string, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "high", "low")
  if err != nil {
    return
  }

  highs := data[0]
  lows := data[1]

  var fvgs []string
  for i := 0; i < len(highs)-3; i++ {
    // Bullish FVG (Gap Up): k1.High < k3.Low
    if highs[i] < lows[i+2] {
      fvgs = append(fvgs, fmt.Sprintf("%.4f,%.4f,1", lows[i+2], highs[i]))
    }
    // Bearish FVG (Gap Down): k1.Low > k3.High
    if lows[i] > highs[i+2] {
      fvgs = append(fvgs, fmt.Sprintf("%.4f,%.4f,2", lows[i], highs[i+2]))
    }
  }

  if len(fvgs) == 0 {
    return
  }

  lastIdx := len(timestamps) - 1
  day, err := r.Day(timestamps[lastIdx-1] / 1000)
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
    "fvg",
    strings.Join(fvgs, ";"),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}
