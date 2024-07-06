package margin

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/margin/isolated"
)

type Isolated struct{}

func NewIsolated() *Isolated {
  return &Isolated{}
}

func (m *Isolated) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &isolated.Order{},
    &isolated.Position{},
    &isolated.Scalping{},
    &isolated.ScalpingPlan{},
    &isolated.Trigger{},
  )
  isolated.NewTradings().AutoMigrate(db)
  return nil
}
