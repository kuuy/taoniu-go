package daily

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/binance/spot/analysis/daily/margin"
)

type Margin struct{}

func NewMargin() *Margin {
	return &Margin{}
}

func (m *Margin) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&margin.Isolated{},
	)
	return nil
}
