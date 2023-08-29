package spot

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/spot/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
  return &Tradings{}
}

func (m *Tradings) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &tradings.Launchpad{},
    &tradings.Scalping{},
    &tradings.Trigger{},
  )
  return nil
}
