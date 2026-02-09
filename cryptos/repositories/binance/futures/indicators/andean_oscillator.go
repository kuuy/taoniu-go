package indicators

import (
  "fmt"
  "math"
  "strconv"
  "strings"
  "time"

  config "taoniu.local/cryptos/config/binance/futures"
)

type AndeanOscillatorRepository struct {
  BaseRepository
}

func (r *AndeanOscillatorRepository) Get(symbol, interval string) (
  bull,
  bear float64,
  price float64,
  timestamp int64,
  err error) {
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
    "andean_oscillator",
  ).Result()
  if err != nil {
    return
  }
  data := strings.Split(val, ",")
  if len(data) < 8 {
    err = fmt.Errorf("invalid data in redis")
    return
  }
  bull, _ = strconv.ParseFloat(data[0], 64)
  bear, _ = strconv.ParseFloat(data[1], 64)
  price, _ = strconv.ParseFloat(data[2], 64)
  timestamp, _ = strconv.ParseInt(data[3], 10, 64)
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

  up1[0] = closes[0]
  up2[0] = math.Pow(closes[0], 2)
  dn1[0] = closes[0]
  dn2[0] = math.Pow(closes[0], 2)

  alpha := 2 / (float64(period) + 1)

  for i := 1; i < len(opens); i++ {
    close2 := math.Pow(closes[i], 2)
    open2 := math.Pow(opens[i], 2)

    up1[i] = math.Max(math.Max(closes[i], opens[i]), up1[i-1]-alpha*(up1[i-1]-closes[i]))
    up2[i] = math.Max(math.Max(close2, open2), up2[i-1]-alpha*(up2[i-1]-close2))
    dn1[i] = math.Min(math.Min(closes[i], opens[i]), dn1[i-1]+alpha*(closes[i]-dn1[i-1]))
    dn2[i] = math.Min(math.Min(close2, open2), dn2[i-1]+alpha*(close2-dn2[i-1]))

    bulls[i] = math.Sqrt(math.Max(0, dn2[i]-math.Pow(dn1[i], 2)))
    bears[i] = math.Sqrt(math.Max(0, up2[i]-math.Pow(up1[i], 2)))
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
    "andean_oscillator",
    fmt.Sprintf(
      "%s,%s,%s,%d",
      strconv.FormatFloat(bulls[lastIdx], 'f', -1, 64),
      strconv.FormatFloat(bears[lastIdx], 'f', -1, 64),
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
