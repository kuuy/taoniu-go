package analysis

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/futures/analysis/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
  return &Tradings{}
}

func (m *Tradings) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &tradings.Scalping{},
  )
  return nil
}
