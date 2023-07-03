package binance

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/futures"
)

type Futures struct{}

func NewFutures() *Futures {
  return &Futures{}
}

func (m *Futures) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &futures.Symbol{},
    &futures.Position{},
    &futures.Order{},
    &futures.Kline{},
    &futures.Strategy{},
    &futures.Plan{},
    &futures.Trigger{},
  )

  futures.NewTradings().AutoMigrate(db)

  return nil
}
