package tradings

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/futures/tradings/triggers"
)

type Triggers struct{}

func NewTriggers() *Triggers {
  return &Triggers{}
}

func (m *Triggers) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &triggers.Grid{},
  )
  return nil
}
