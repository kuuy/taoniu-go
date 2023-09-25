package futures

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/futures/patterns"
)

type Patterns struct{}

func NewPatterns() *Patterns {
  return &Patterns{}
}

func (m *Patterns) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &patterns.Candlesticks{},
  )
  return nil
}
