package indicators

import (
  "fmt"
  "strconv"
  "time"

  "github.com/markcheno/go-talib"

  config "taoniu.local/cryptos/config/binance/futures"
)

type VolumeMovingRepository struct {
  BaseRepository
}

func (r *VolumeMovingRepository) Get(symbol, interval string) (
  volumeMoving float64,
  volumeRatio float64,
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
    "volume_moving",
    "volume_ratio",
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
    case "volume_moving":
      volumeMoving, _ = strconv.ParseFloat(data[i].(string), 64)
    case "volume_ratio":
      volumeRatio, _ = strconv.ParseFloat(data[i].(string), 64)
    }
  }
  return
}

func (r *VolumeMovingRepository) Flush(symbol string, interval string, period int, limit int) (err error) {
  data, timestamps, err := r.Klines(symbol, interval, limit, "volume")
  if err != nil {
    return
  }

  volumes := data[0]
  lastIdx := len(timestamps) - 1

  result := talib.Sma(volumes, period)

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
  r.Rdb.HMSet(
    r.Ctx,
    redisKey,
    map[string]interface{}{
      "volume_moving": strconv.FormatFloat(result[len(result)-1], 'f', -1, 64),
      "volume_ratio":  strconv.FormatFloat(volumes[len(result)-1]/result[len(result)-1], 'f', -1, 64),
    },
  )
  ttl, _ := r.Rdb.TTL(r.Ctx, redisKey).Result()
  if -1 == ttl.Nanoseconds() {
    r.Rdb.Expire(r.Ctx, redisKey, time.Hour*24)
  }

  return nil
}
