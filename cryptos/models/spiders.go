package models

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/spiders"
)

type Spiders struct{}

func NewSpiders() *Spiders {
  return &Spiders{}
}

func (m *Spiders) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &spiders.Source{},
  )
  return nil
}
