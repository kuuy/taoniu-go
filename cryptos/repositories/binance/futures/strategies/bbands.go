package strategies

import (
  "errors"

  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type BBandsRepository struct {
  BaseRepository
  Repository *repositories.BBandsRepository
}

func (r *BBandsRepository) Flush(symbol string, interval string) (err error) {
  b1, b2, b3, w1, w2, w3, price, timestamp, err := r.Repository.Get(symbol, interval)
  if err != nil {
    return
  }

  var signal int
  if b1 < 0.5 && b2 < 0.5 && b3 > 0.5 {
    signal = 1
  }
  if b1 > 0.5 && b2 < 0.5 && b3 < 0.5 {
    signal = 2
  }
  if b1 > 0.8 && b2 > 0.8 && b3 > 0.8 {
    signal = 1
  }
  if b1 > 0.8 && b2 > 0.8 && b3 < 0.8 {
    signal = 2
  }
  if w1 < 0.2 && w2 < 0.2 && w3 < 0.2 {
    if w1 < 0.06 || w2 < 0.06 || w3 > 0.06 {
      return nil
    }
  }
  if signal == 0 {
    return
  }

  var entity models.Strategy
  result := r.Db.Where(
    "symbol=? AND indicator=? AND interval=?",
    symbol,
    "bbands",
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
    Indicator: "bbands",
    Interval:  interval,
    Price:     price,
    Signal:    signal,
    Timestamp: timestamp,
  }
  r.Db.Create(&entity)
  return
}
