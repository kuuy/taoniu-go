package indicators

import (
  "context"
  "errors"
  "fmt"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
)

type AndeanOscillatorData struct {
  Bull      float64
  Bear      float64
  Timestamp int64
}

type BaseRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *BaseRepository) Day(timestamp int64) (day string, err error) {
  now := time.Now()
  last := time.Unix(timestamp, 0)
  if now.UTC().Format("0102") != last.UTC().Format("0102") {
    err = errors.New("timestamp is not today")
    return
  }
  day = now.Format("0102")
  return
}

func (r *BaseRepository) Timestep(interval string) int64 {
  switch interval {
  case "1m":
    return 60000
  case "15m":
    return 900000
  case "4h":
    return 14400000
  }
  return 86400000
}

func (r *BaseRepository) Timestamp(interval string) int64 {
  now := time.Now().UTC()
  duration := -time.Second * time.Duration(now.Second())
  switch interval {
  case "15m":
    minute, _ := decimal.NewFromInt(int64(now.Minute())).Div(decimal.NewFromInt(15)).Floor().Mul(decimal.NewFromInt(15)).Float64()
    duration = duration - time.Minute*time.Duration(now.Minute()-int(minute))
  case "4h":
    hour, _ := decimal.NewFromInt(int64(now.Hour())).Div(decimal.NewFromInt(4)).Floor().Mul(decimal.NewFromInt(4)).Float64()
    duration = duration - time.Hour*time.Duration(now.Hour()-int(hour)) - time.Minute*time.Duration(now.Minute())
  case "1d":
    duration = duration - time.Hour*time.Duration(now.Hour()) - time.Minute*time.Duration(now.Minute())
  }
  return now.Add(duration).Unix() * 1000
}

func (r *BaseRepository) Klines(symbol, interval string, limit int, fields ...string) (result [][]float64, timestamps []int64, err error) {
  var klines []*models.Kline
  err = r.Db.Select(
    append(fields, "timestamp"),
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Limit(
    limit,
  ).Find(
    &klines,
  ).Error
  if err != nil {
    return
  }
  if len(klines) < limit {
    err = fmt.Errorf("klines not enough")
    return
  }

  result = make([][]float64, len(fields))
  for i := range fields {
    result[i] = make([]float64, limit)
  }
  timestamps = make([]int64, limit)

  timestep := r.Timestep(interval)
  for i, item := range klines {
    pos := limit - i - 1
    if i > 0 && (klines[i-1].Timestamp-item.Timestamp) != timestep {
      err = fmt.Errorf("[%s] %s klines lost", symbol, interval)
      return
    }

    for j, field := range fields {
      var val float64
      switch field {
      case "open":
        val = item.Open
      case "high":
        val = item.High
      case "low":
        val = item.Low
      case "close":
        val = item.Close
      case "volume":
        val = item.Volume
      }
      result[j][pos] = val
    }

    timestamps[pos] = item.Timestamp
  }
  return
}
