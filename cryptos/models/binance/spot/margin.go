package spot

import (
  "gorm.io/gorm"
  models "taoniu.local/cryptos/models/binance/spot/margin"
)

type Margin struct{}

func NewMargin() *Margin {
  return &Margin{}
}

func (m *Margin) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &models.Order{},
  )
  models.NewCross().AutoMigrate(db)
  models.NewIsolated().AutoMigrate(db)
  return nil
}
