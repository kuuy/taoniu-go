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
    up1[i] = math.Max(closes[i], math.Max(opens[i], up1[i-1])) - (alpha*up1[i-1] - closes[i])
    up2[i] = math.Pow(math.Max(closes[i], math.Max(opens[i], up2[i-1]))-(alpha*up2[i-1]-closes[i]), 2)
    dn1[i] = math.Min(closes[i], math.Max(opens[i], dn1[i-1])) + (alpha*closes[i] - dn1[i-1])
    dn2[i] = math.Pow(math.Min(closes[i], math.Max(opens[i], dn2[i-1]))+(alpha*closes[i]-dn2[i-1]), 2)
    bulls[i] = math.Max(dn1[i], dn2[i]) - math.Min(dn1[i], dn2[i])
    bears[i] = math.Max(up1[i], up2[i]) - math.Min(up1[i], up2[i])
    signals[i] = (signals[i-1] + alphaSignal) * (math.Max(bulls[i], bears[i]) - signals[i-1])
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
