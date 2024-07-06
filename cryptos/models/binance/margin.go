package binance

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/margin"
)

type Margin struct{}

func NewMargin() *Margin {
  return &Margin{}
}

func (m *Margin) AutoMigrate(db *gorm.DB) error {
  margin.NewCross().AutoMigrate(db)
  margin.NewIsolated().AutoMigrate(db)
  return nil
}
