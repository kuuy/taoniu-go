package indicators

import (
  "fmt"
  "math"
  "strconv"
  "time"

  config "taoniu.local/cryptos/config/binance/futures"
)

type AndeanOscillatorRepository struct {
  BaseRepository
}

func (r *AndeanOscillatorRepository) Get(symbol, interval string) (
  bull,
  bear,
  signal float64,
  err error) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )

  fields := []string{
    "ao_bull",
    "ao_bear",
    "ao_signal",
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
    case "ao_bull":
      bull, _ = strconv.ParseFloat(data[i].(string), 64)
    case "ao_bear":
      bear, _ = strconv.ParseFloat(data[i].(string), 64)
    case "ao_signal":
      signal, _ = strconv.ParseFloat(data[i].(string), 64)
    }
  }
  return
}

func (r *AndeanOscillatorRepository) Flush(symbol string, interval string, period int, length int, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "open", "close")
  if err != nil {
    return
  }

  opens := data[0]
  closes := data[1]
  lastIdx := len(timestamps) - 1

  up1 := make([]float64, limit)
  up2 := make([]float64, limit)
  dn1 := make([]float64, limit)
  dn2 := make([]float64, limit)
  bulls := make([]float64, limit)
  bears := make([]float64, limit)
  signals := make([]float64, limit)

  up1[0] = closes[0]
  up2[0] = math.Pow(closes[0], 2)
  dn1[0] = closes[0]
  dn2[0] = math.Pow(closes[0], 2)
  signals[0] = closes[0]

  alpha := 2 / (float64(period) + 1)
  alphaSignal := 2 / (float64(length) + 1)

  for i := 1; i < len(opens); i++ {
    close2 := math.Pow(closes[i], 2)
    open2 := math.Pow(opens[i], 2)

    up1[i] = math.Max(math.Max(closes[i], opens[i]), up1[i-1]-alpha*(up1[i-1]-closes[i]))
    up2[i] = math.Max(math.Max(close2, open2), up2[i-1]-alpha*(up2[i-1]-close2))
    dn1[i] = math.Min(math.Min(closes[i], opens[i]), dn1[i-1]-alpha*(dn1[i-1]-closes[i]))
    dn2[i] = math.Min(math.Min(close2, open2), dn2[i-1]-alpha*(dn2[i-1]-close2))

    bulls[i] = math.Sqrt(math.Max(math.Pow(up1[i], 2)-up2[i], 0))
    bears[i] = math.Sqrt(math.Max(dn2[i]-math.Pow(dn1[i], 2), 0))

    signals[i] = signals[i-1] + alphaSignal*(math.Max(bulls[i], bears[i])-signals[i-1])
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
  r.Rdb.HMSet(
    r.Ctx,
    redisKey,
    map[string]interface{}{
      "ao_bull":   bulls[lastIdx],
      "ao_bear":   bears[lastIdx],
      "ao_signal": signals[lastIdx],
    },
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}
