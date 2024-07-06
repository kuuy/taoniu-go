package cross

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/margin/cross/analysis/tradings"
)

type Analysis struct{}

func NewAnalysis() *Analysis {
  return &Analysis{}
}

func (m *Analysis) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &tradings.Scalping{},
    &tradings.Trigger{},
  )
  return nil
}
