package indicators

import (
  "fmt"
  "github.com/markcheno/go-talib"
  "strconv"
  "time"

  config "taoniu.local/cryptos/config/binance/futures"
)

type AtrRepository struct {
  BaseRepository
}

func (r *AtrRepository) Get(symbol, interval string) (result float64, err error) {
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
    "atr",
  ).Result()
  if err != nil {
    return
  }
  result, err = strconv.ParseFloat(val, 64)
  return
}

func (r *AtrRepository) Flush(symbol string, interval string, period int, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "close", "high", "low")
  if err != nil {
    return
  }

  prices := data[0]
  highs := data[1]
  lows := data[2]
  lastIdx := len(timestamps) - 1

  result := talib.Atr(
    highs,
    lows,
    prices,
    period,
  )

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
    "atr",
    strconv.FormatFloat(result[limit-1], 'f', -1, 64),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}

func (r *AtrRepository) Multiplier(price, atr float64) float64 {
  if price == 0 {
    return 2.0
  }
  volatility := atr / price
  switch {
  case volatility > 0.05: // 高波动 >5%
    return 2.5
  case volatility > 0.03: // 中波动 3-5%
    return 2.0
  case volatility > 0.015: // 中低波动 1.5-3%
    return 1.5
  case volatility > 0.008: // 低波动 0.8-1.5%
    return 1.2
  default: // 极低波动 <0.8%
    return 1.0
  }
}
