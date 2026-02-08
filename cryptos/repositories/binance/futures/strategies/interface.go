package strategies

import (
  "context"
  "errors"
  "fmt"
  "strconv"
  "strings"
  config "taoniu.local/cryptos/config/binance/futures"
  "time"

  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"
  models "taoniu.local/cryptos/models/binance/futures"
)

type BaseRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *BaseRepository) Price(symbol string) (price float64, err error) {
  val, err := r.Rdb.HGet(
    r.Ctx,
    fmt.Sprintf(config.REDIS_KEY_TICKERS, symbol),
    "price",
  ).Result()
  if err != nil {
    return
  }
  price, _ = strconv.ParseFloat(val, 64)
  return
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
    minute := float64(now.Minute() / 15 * 15)
    duration = duration - time.Minute*time.Duration(now.Minute()-int(minute))
  case "4h":
    hour := float64(now.Hour() / 4 * 4)
    duration = duration - time.Hour*time.Duration(now.Hour()-int(hour)) - time.Minute*time.Duration(now.Minute())
  case "1d":
    duration = duration - time.Hour*time.Duration(now.Hour()) - time.Minute*time.Duration(now.Minute())
  }
  return now.Add(duration).Unix() * 1000
}
