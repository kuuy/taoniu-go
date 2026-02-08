package strategies

import (
  "errors"

  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type KdjRepository struct {
  BaseRepository
  Repository *repositories.KdjRepository
}

func (r *KdjRepository) Flush(symbol string, interval string) (err error) {
  slowk, slowd, slowj, price, timestamp, err := r.Repository.Get(symbol, interval)
  if err != nil {
    return
  }

  var signal int
  if slowk < 20 && slowd < 30 && slowj < 60 {
    signal = 1
  }
  if slowk > 80 && slowd > 70 && slowj > 90 {
    signal = 2
  }

  if signal == 0 {
    return
  }

  var entity models.Strategy
  result := r.Db.Where(
    "symbol=? AND indicator=? AND interval=?",
    symbol,
    "kdj",
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
    Indicator: "kdj",
    Interval:  interval,
    Price:     price,
    Signal:    signal,
    Timestamp: timestamp,
  }
  r.Db.Create(&entity)
  return
}
