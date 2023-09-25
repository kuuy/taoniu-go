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
    &futures.Kline{},
    &futures.Strategy{},
    &futures.Plan{},
    &futures.Order{},
    &futures.Position{},
    &futures.Scalping{},
    &futures.ScalpingPlan{},
    &futures.Trigger{},
  )

  futures.NewPatterns().AutoMigrate(db)
  futures.NewTradings().AutoMigrate(db)
  futures.NewAnalysis().AutoMigrate(db)

  return nil
}
