package margin

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/spot/analysis/margin/isolated/tradings"
)

type Isolated struct{}

func NewIsolated() *Isolated {
  return &Isolated{}
}

func (m *Isolated) AutoMigrate(db *gorm.DB) error {
  tradings.NewFishers().AutoMigrate(db)
  return nil
}
