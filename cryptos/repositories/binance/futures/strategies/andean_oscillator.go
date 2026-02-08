package strategies

import (
  "context"
  "errors"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type AndeanOscillatorRepository struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.AndeanOscillatorRepository
}

func (r *AndeanOscillatorRepository) Flush(symbol string, interval string) (err error) {
  bull, bear, _, err := r.Repository.Get(symbol, interval)
  if err != nil {
    return
  }

  var signal int
  if bull > bear {
    signal = 1
  } else if bear > bull {
    signal = 2
  }

  if signal == 0 {
    return
  }

  // Fetch current price (needed for strategy record)
  // Since Get() doesn't return price or timestamp, we might need to fetch it separately or assume current
  // But ideally we should use the timestamp from the indicator or kline.
  // The AndeanOscillator Get() in indicators/andean_oscillator.go ONLY returns bull, bear, signal.
  // It does NOT return timestamp or price.
  // We need to fetch the latest kline to get price and timestamp.
  var kline models.Kline
  err = r.Db.Select("close", "timestamp").Where("symbol = ? AND interval = ?", symbol, interval).Order("timestamp desc").First(&kline).Error
  if err != nil {
    return
  }
  price := kline.Close
  timestamp := kline.Timestamp

  // Use kline timestamp for consistency
  // Check if strategy already exists
  var entity models.Strategy
  result := r.Db.Where(
    "symbol=? AND indicator=? AND interval=?",
    symbol,
    "andean_oscillator",
    interval,
  ).Order(
    "timestamp DESC",
  ).Take(&entity)

  if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
    if entity.Signal == signal {
      return
    }
    if entity.Timestamp >= timestamp {
      return
    }
  }

  entity = models.Strategy{
    ID:        xid.New().String(),
    Symbol:    symbol,
    Indicator: "andean_oscillator",
    Interval:  interval,
    Price:     price,
    Signal:    signal,
    Timestamp: timestamp,
  }
  r.Db.Create(&entity)
  return
}
