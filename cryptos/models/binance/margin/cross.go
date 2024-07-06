package margin

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/margin/cross"
)

type Cross struct{}

func NewCross() *Cross {
  return &Cross{}
}

func (m *Cross) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &cross.Order{},
    &cross.Position{},
    &cross.Scalping{},
    &cross.ScalpingPlan{},
    &cross.Trigger{},
  )
  cross.NewTradings().AutoMigrate(db)
  cross.NewAnalysis().AutoMigrate(db)
  return nil
}
