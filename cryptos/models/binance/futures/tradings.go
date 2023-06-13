package futures

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/futures/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
  return &Tradings{}
}

func (m *Tradings) AutoMigrate(db *gorm.DB) error {
  tradings.NewTriggers().AutoMigrate(db)
  return nil
}
