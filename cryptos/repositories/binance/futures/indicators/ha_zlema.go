package indicators

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/markcheno/go-talib"

	config "taoniu.local/cryptos/config/binance/futures"
)

type HaZlemaRepository struct {
  BaseRepository
}

func (r *HaZlemaRepository) Get(symbol, interval string) (
  prev,
  current,
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
    "ha_zlema",
  ).Result()
  if err != nil {
    return
  }
  data := strings.Split(val, ",")
  if len(data) < 4 {
    err = fmt.Errorf("invalid data in redis")
    return
  }
  prev, _ = strconv.ParseFloat(data[0], 64)
  current, _ = strconv.ParseFloat(data[1], 64)
  price, _ = strconv.ParseFloat(data[2], 64)
  timestamp, _ = strconv.ParseInt(data[3], 10, 64)
  return
}

func (r *HaZlemaRepository) Flush(symbol string, interval string, period int, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "open", "close", "high", "low")
  if err != nil {
    return
  }

  opens := data[0]
  closes := data[1]
  highs := data[2]
  lows := data[3]
  lastIdx := len(timestamps) - 1

  // 1. Calculate Heiken Ashi
  haClose := make([]float64, len(opens))
  haOpen := make([]float64, len(opens))
  haHigh := make([]float64, len(opens))
  haLow := make([]float64, len(opens))

  for i := 0; i < len(opens); i++ {
    haClose[i] = (opens[i] + closes[i] + highs[i] + lows[i]) / 4
    if i == 0 {
      haOpen[i] = (opens[0] + closes[0]) / 2
    } else {
      haOpen[i] = (haOpen[i-1] + haClose[i-1]) / 2
    }
    haHigh[i] = math.Max(highs[i], math.Max(haOpen[i], haClose[i]))
    haLow[i] = math.Min(lows[i], math.Min(haOpen[i], haClose[i]))
  }

  // 2. HA typical price
  haTypical := make([]float64, len(haOpen))
  for i := 0; i < len(haOpen); i++ {
    haTypical[i] = (haOpen[i] + haClose[i] + haHigh[i] + haLow[i]) / 4
  }

  // 3. ZLEMA
  lag := period - 1
  zdata := make([]float64, len(haTypical))
  for i := lag; i < len(haTypical); i++ {
    zdata[i] = haTypical[i] + (haTypical[i] - haTypical[i-lag])
  }

  result := talib.Ema(zdata, period)
  if len(result) < 2 {
    return fmt.Errorf("talib calculation failed to return enough data")
  }

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
    "ha_zlema",
    fmt.Sprintf(
      "%v,%v,%v,%d",
      strconv.FormatFloat(result[len(result)-2], 'f', -1, 64),
      strconv.FormatFloat(result[len(result)-1], 'f', -1, 64),
      strconv.FormatFloat(closes[lastIdx], 'f', -1, 64),
      timestamps[lastIdx],
    ),
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}
