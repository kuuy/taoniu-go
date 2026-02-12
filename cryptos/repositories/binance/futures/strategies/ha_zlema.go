package strategies

import (
  "errors"

  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type HaZlemaRepository struct {
  BaseRepository
  Repository *repositories.HaZlemaRepository
}

func (r *HaZlemaRepository) Flush(symbol string, interval string) (err error) {
  prev, current, price, timestamp, err := r.Repository.Get(symbol, interval)
  if err != nil {
    return
  }

  if prev*current >= 0.0 {
    return
  }

  var signal int
  if price > current {
    signal = 1
  } else if price < current {
    signal = 2
  }

  if signal == 0 {
    return
  }

  var entity models.Strategy
  result := r.Db.Where(
    "symbol=? AND indicator=? AND interval=?",
    symbol,
    "ha_zlema",
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
    Indicator: "ha_zlema",
    Interval:  interval,
    Price:     price,
    Signal:    signal,
    Timestamp: timestamp,
  }
  r.Db.Create(&entity)
  return
}
