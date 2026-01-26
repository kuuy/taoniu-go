package models

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/raydium"
)

type Raydium struct{}

func NewRaydium() *Raydium {
  return &Raydium{}
}

func (m *Raydium) AutoMigrate(db *gorm.DB) error {
  raydium.NewSwap().AutoMigrate(db)
  return nil
}
