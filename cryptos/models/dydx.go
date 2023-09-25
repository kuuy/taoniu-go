package models

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/dydx"
)

type Dydx struct{}

func NewDydx() *Dydx {
  return &Dydx{}
}

func (m *Dydx) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &dydx.Market{},
    &dydx.Kline{},
    &dydx.Strategy{},
    &dydx.Plan{},
    &dydx.Order{},
    &dydx.Position{},
    &dydx.Scalping{},
    &dydx.ScalpingPlan{},
    &dydx.Trigger{},
  )

  dydx.NewPatterns().AutoMigrate(db)
  dydx.NewTradings().AutoMigrate(db)

  return nil
}
