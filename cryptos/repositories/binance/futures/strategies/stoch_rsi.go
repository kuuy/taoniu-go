package strategies

import (
  "errors"

  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type StochRsiRepository struct {
  BaseRepository
  Repository *repositories.StochRsiRepository
}

func (r *StochRsiRepository) Flush(symbol string, interval string) (err error) {
  fastk, fastd, price, timestamp, err := r.Repository.Get(symbol, interval)
  if err != nil {
    return
  }

  var signal int
  if fastk < 20 && fastd < 30 {
    signal = 1
  }
  if fastk > 80 && fastd > 70 {
    signal = 2
  }

  if signal == 0 {
    return
  }

  var entity models.Strategy
  result := r.Db.Where(
    "symbol=? AND indicator=? AND interval=?",
    symbol,
    "rsi_stoch",
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
    Indicator: "rsi_stoch",
    Interval:  interval,
    Price:     price,
    Signal:    signal,
    Timestamp: timestamp,
  }
  r.Db.Create(&entity)
  return
}
