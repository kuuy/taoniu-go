package indicators

import (
  "fmt"
  "strconv"
  "strings"
  "time"

  "github.com/markcheno/go-talib"
  config "taoniu.local/cryptos/config/binance/futures"
)

type StochRsiRepository struct {
  BaseRepository
}

type StochRsiData struct {
  StochK    float64
  StochD    float64
  Timestamp int64
}

func (r *StochRsiRepository) Get(symbol, interval string) (result *StochRsiData, err error) {
  day := time.Now().Format("0102")
  redisKey := fmt.Sprintf(
    config.REDIS_KEY_INDICATORS,
    interval,
    symbol,
    day,
  )
  val, err := r.Rdb.HGet(r.Ctx, redisKey, "stoch_rsi").Result()
  if err != nil {
    return nil, err
  }
  parts := strings.Split(val, ",")
  if len(parts) < 3 {
    return nil, fmt.Errorf("invalid stoch_rsi data")
  }
  k, _ := strconv.ParseFloat(parts[0], 64)
  d, _ := strconv.ParseFloat(parts[1], 64)
  ts, _ := strconv.ParseInt(parts[2], 10, 64)
  return &StochRsiData{
    StochK:    k,
    StochD:    d,
    Timestamp: ts,
  }, nil
}

func (r *StochRsiRepository) Flush(symbol string, interval string, period int, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "close")
  if err != nil {
    return
  }

  closes := data[0]
  lastIdx := len(timestamps) - 1

  rsi := talib.Rsi(closes, period)
  fastk, fastd := talib.Stoch(rsi, rsi, rsi, 14, 3, talib.SMA, 3, talib.SMA)

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
    "stoch_rsi",
    fmt.Sprintf(
      "%v,%v,%d",
      fastk[lastIdx],
      fastd[lastIdx],
      timestamps[lastIdx],
    ),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return
}
