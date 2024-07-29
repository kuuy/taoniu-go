package analysis

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/spot/analysis/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
  return &Tradings{}
}

func (m *Tradings) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &tradings.Scalping{},
    &tradings.Trigger{},
  )
  return nil
}
