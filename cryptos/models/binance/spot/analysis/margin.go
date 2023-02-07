package analysis

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/binance/spot/analysis/margin"
)

type Margin struct{}

func NewMargin() *Margin {
	return &Margin{}
}

func (m *Margin) AutoMigrate(db *gorm.DB) error {
	margin.NewIsolated().AutoMigrate(db)
	return nil
}
