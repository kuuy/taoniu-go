package indicators

import (
  "context"
  "errors"
  "fmt"
  "strconv"
  "strings"
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

type VolumeSegment struct {
  MinPrice float64
  MaxPrice float64
  Volume   float64
}

type BaseRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *BaseRepository) Filters(symbol string) (tickSize float64, stepSize float64, err error) {
  var entity *models.Symbol
  err = r.Db.Select("filters").Where("symbol", symbol).Take(&entity).Error
  if err != nil {
    return
  }

  var values []string
  values = strings.Split(entity.Filters["price"].(string), ",")
  tickSize, _ = strconv.ParseFloat(values[2], 64)
  values = strings.Split(entity.Filters["quote"].(string), ",")
  stepSize, _ = strconv.ParseFloat(values[2], 64)
  return
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

func (r *BaseRepository) Kline(symbol, interval string, fields ...string) (data []float64, timestamp int64, err error) {
  var kline *models.Kline
  err = r.Db.Select(
    append(fields, "timestamp"),
  ).Where(
    "symbol=? AND interval=?", symbol, interval,
  ).Order(
    "timestamp desc",
  ).Take(
    &kline,
  ).Error
  if err != nil {
    return
  }

  if kline.Timestamp < r.Timestamp(interval)-60000 {
    err = fmt.Errorf("[%s] waiting for %s klines flush", symbol, interval)
    return
  }

  data = make([]float64, len(fields))
  for i, field := range fields {
    var val float64
    switch field {
    case "open":
      val = kline.Open
    case "high":
      val = kline.High
    case "low":
      val = kline.Low
    case "close":
      val = kline.Close
    case "volume":
      val = kline.Volume
    }
    data[i] = val
  }
  timestamp = kline.Timestamp
  return
}

func (r *BaseRepository) Klines(symbol, interval string, limit int, fields ...string) (data [][]float64, timestamps []int64, err error) {
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

  if klines[0].Timestamp < r.Timestamp(interval)-60000 {
    err = fmt.Errorf("[%s] waiting for %s klines flush", symbol, interval)
    return
  }

  data = make([][]float64, len(fields))
  for i := range fields {
    data[i] = make([]float64, limit)
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
      data[j][pos] = val
    }

    timestamps[pos] = item.Timestamp
  }
  return
}
