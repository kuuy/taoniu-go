package futures

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/futures/analysis/tradings"
)

type Analysis struct{}

func NewAnalysis() *Analysis {
  return &Analysis{}
}

func (m *Analysis) AutoMigrate(db *gorm.DB) error {
  db.AutoMigrate(
    &tradings.Scalping{},
  )
  return nil
}
