package models

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance"
)

type Binance struct{}

func NewBinance() *Binance {
  return &Binance{}
}

func (m *Binance) AutoMigrate(db *gorm.DB) error {
  binance.NewSpot().AutoMigrate(db)
  binance.NewMargin().AutoMigrate(db)
  binance.NewFutures().AutoMigrate(db)
  binance.NewSavings().AutoMigrate(db)
  return nil
}
