package margin

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/spot/margin/cross"
)

type Cross struct{}

func NewCross() *Cross {
  return &Cross{}
}

func (m *Cross) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &cross.Trigger{},
  )
  cross.NewTradings().AutoMigrate(db)
  return nil
}
