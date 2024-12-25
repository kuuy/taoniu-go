package spot

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/spot/gambling"
)

type Gambling struct{}

func NewGambling() *Gambling {
  return &Gambling{}
}

func (m *Gambling) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &gambling.Ant{},
  )
  return nil
}
