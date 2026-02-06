package indicators

import (
  "errors"
  "fmt"
  "math"
  "strconv"
  "strings"
  "time"

  "github.com/shopspring/decimal"

  config "taoniu.local/cryptos/config/binance/futures"
)

type AndeanOscillatorRepository struct {
  BaseRepository
}

func (r *AndeanOscillatorRepository) Get(symbol, interval string) (*AndeanOscillatorData, error) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  val, err := r.Rdb.HGet(r.Ctx, redisKey, "andean").Result()
  if err != nil {
    return nil, err
  }
  parts := strings.Split(val, ",")
  if len(parts) < 3 {
    return nil, fmt.Errorf("invalid andean data")
  }
  bull, _ := strconv.ParseFloat(parts[0], 64)
  bear, _ := strconv.ParseFloat(parts[1], 64)
  ts, _ := strconv.ParseInt(parts[2], 10, 64)
  return &AndeanOscillatorData{
    Bull:      bull,
    Bear:      bear,
    Timestamp: ts,
  }, nil
}

func (r *AndeanOscillatorRepository) Flush(symbol string, interval string, period int, length int, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "open", "close")
  if err != nil {
    return
  }

  opens := data[0]
  closes := data[0]

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

  alpha, _ := decimal.NewFromInt(2).Div(decimal.NewFromInt(int64(period + 1))).Float64()
  alphaSignal, _ := decimal.NewFromInt(2).Div(decimal.NewFromInt(int64(length + 1))).Float64()

  for i := 1; i < len(opens); i++ {
    up1[i], _ = decimal.Max(
      decimal.NewFromFloat(closes[i]),
      decimal.NewFromFloat(opens[i]),
      decimal.NewFromFloat(up1[i-1]).Sub(decimal.NewFromFloat(alpha).Mul(decimal.NewFromFloat(up1[i-1]).Sub(decimal.NewFromFloat(closes[i])))),
    ).Float64()
    up2[i], _ = decimal.Max(
      decimal.NewFromFloat(closes[i]).Pow(decimal.NewFromInt(2)),
      decimal.NewFromFloat(opens[i]).Pow(decimal.NewFromInt(2)),
      decimal.NewFromFloat(up2[i-1]).Sub(decimal.NewFromFloat(alpha).Mul(decimal.NewFromFloat(up2[i-1]).Sub(decimal.NewFromFloat(closes[i]).Pow(decimal.NewFromInt(2))))),
    ).Float64()
    dn1[i], _ = decimal.Min(
      decimal.NewFromFloat(closes[i]),
      decimal.NewFromFloat(opens[i]),
      decimal.NewFromFloat(dn1[i-1]).Add(decimal.NewFromFloat(alpha).Mul(decimal.NewFromFloat(closes[i]).Sub(decimal.NewFromFloat(dn1[i-1])))),
    ).Float64()
    dn2[i], _ = decimal.Min(
      decimal.NewFromFloat(closes[i]).Pow(decimal.NewFromInt(2)),
      decimal.NewFromFloat(opens[i]).Pow(decimal.NewFromInt(2)),
      decimal.NewFromFloat(dn2[i-1]).Add(decimal.NewFromFloat(alpha).Mul(decimal.NewFromFloat(closes[i]).Pow(decimal.NewFromInt(2)).Sub(decimal.NewFromFloat(dn2[i-1])))),
    ).Float64()
    bulls[i], _ = decimal.NewFromFloat(dn2[i]).Sub(decimal.NewFromFloat(dn1[i]).Pow(decimal.NewFromInt(2))).Float64()
    bears[i], _ = decimal.NewFromFloat(up2[i]).Sub(decimal.NewFromFloat(up1[i]).Pow(decimal.NewFromInt(2))).Float64()
    if bulls[i] < 0 || bears[i] < 0 {
      err = errors.New("calc Andean Oscillator Failed")
      return
    }
    bulls[i] = math.Sqrt(bulls[i])
    bears[i] = math.Sqrt(bears[i])
    signals[i], _ = decimal.NewFromFloat(signals[i-1]).Add(decimal.NewFromFloat(alphaSignal).Mul(decimal.Max(
      decimal.NewFromFloat(bulls[i]),
      decimal.NewFromFloat(bears[i]),
    ).Sub(decimal.NewFromFloat(signals[i-1])))).Float64()
  }

  day, err := r.Day(timestamps[0] / 1000)
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
      "ao_bull":   bulls[len(opens)-1],
      "ao_bear":   bears[len(opens)-1],
      "ao_signal": signals[len(opens)-1],
    },
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}
