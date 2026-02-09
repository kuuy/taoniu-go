package strategies

import (
  "errors"

  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type AndeanOscillatorRepository struct {
  BaseRepository
  Repository *repositories.AndeanOscillatorRepository
}

func (r *AndeanOscillatorRepository) Flush(symbol, interval string) (err error) {
  bull, bear, price, timestamp, err := r.Repository.Get(symbol, interval)
  if err != nil {
    return
  }

  var signal int
  if bull > bear {
    signal = 1
  } else if bull < bear {
    signal = 2
  }

  if signal == 0 {
    return
  }

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
