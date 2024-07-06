package spot

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/spot/analysis"
)

type Analysis struct{}

func NewAnalysis() *Analysis {
  return &Analysis{}
}

func (m *Analysis) AutoMigrate(db *gorm.DB) error {
  analysis.NewTradings().AutoMigrate(db)
  return nil
}
