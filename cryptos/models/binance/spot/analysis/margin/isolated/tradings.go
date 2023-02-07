package isolated

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/spot/analysis/margin/isolated/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
	return &Tradings{}
}

func (m *Tradings) AutoMigrate(db *gorm.DB) error {
	tradings.NewFishers().AutoMigrate(db)
	return nil
}
