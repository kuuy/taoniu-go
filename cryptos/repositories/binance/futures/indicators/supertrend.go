package indicators

import (
  "fmt"
  "strconv"
  "strings"
  "time"

  "github.com/markcheno/go-talib"

  config "taoniu.local/cryptos/config/binance/futures"
)

type SuperTrendRepository struct {
  BaseRepository
}

func (r *SuperTrendRepository) Get(symbol, interval string) (
  signal int,
  superTrend,
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
    "supertrend",
  ).Result()
  if err != nil {
    return
  }
  data := strings.Split(val, ",")
  if len(data) < 4 {
    err = fmt.Errorf("invalid data in redis")
    return
  }
  signal, _ = strconv.Atoi(data[0])
  superTrend, _ = strconv.ParseFloat(data[1], 64)
  price, _ = strconv.ParseFloat(data[2], 64)
  timestamp, _ = strconv.ParseInt(data[3], 10, 64)
  return
}

func (r *SuperTrendRepository) Flush(symbol string, interval string, period int, multiplier float64, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "high", "low", "close")
  if err != nil {
    return
  }

  highs := data[0]
  lows := data[1]
  closes := data[2]
  lastIdx := len(timestamps) - 1

  atr := talib.Atr(highs, lows, closes, period)

  upperBands := make([]float64, len(closes))
  lowerBands := make([]float64, len(closes))
  superTrend := make([]float64, len(closes))
  signals := make([]int, len(closes))

  for i := period; i < len(closes); i++ {
    medianPrice := (highs[i] + lows[i]) / 2
    basicUpperBand := medianPrice + (multiplier * atr[i])
    basicLowerBand := medianPrice - (multiplier * atr[i])

    if i == period {
      upperBands[i] = basicUpperBand
      lowerBands[i] = basicLowerBand
      continue
    }

    if basicUpperBand < upperBands[i-1] || closes[i-1] > upperBands[i-1] {
      upperBands[i] = basicUpperBand
    } else {
      upperBands[i] = upperBands[i-1]
    }

    if basicLowerBand > lowerBands[i-1] || closes[i-1] < lowerBands[i-1] {
      lowerBands[i] = basicLowerBand
    } else {
      lowerBands[i] = lowerBands[i-1]
    }

    if superTrend[i-1] == upperBands[i-1] {
      if closes[i] > upperBands[i] {
        superTrend[i] = lowerBands[i]
        signals[i] = 1 // Long
      } else {
        superTrend[i] = upperBands[i]
        signals[i] = 2 // Short
      }
    } else {
      if closes[i] < lowerBands[i] {
        superTrend[i] = upperBands[i]
        signals[i] = 2 // Short
      } else {
        superTrend[i] = lowerBands[i]
        signals[i] = 1 // Long
      }
    }
  }

  day, err := r.Day(timestamps[lastIdx] / 1000)
  if err != nil {
    return
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
    "supertrend",
    fmt.Sprintf(
      "%d,%s,%s,%d",
      signals[lastIdx],
      strconv.FormatFloat(superTrend[lastIdx], 'f', -1, 64),
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
