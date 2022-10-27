package analysis

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/binance/spot/analysis/daily"
)

type Daily struct{}

func NewDaily() *Daily {
	return &Daily{}
}

func (m *Daily) AutoMigrate(db *gorm.DB) error {
	daily.NewMargin().AutoMigrate(db)
	return nil
}
