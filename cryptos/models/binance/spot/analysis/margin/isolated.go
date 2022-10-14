package margin

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/binance/spot/analysis/margin/isolated"
)

type Isolated struct{}

func NewIsolated() *Isolated {
	return &Isolated{}
}

func (m *Isolated) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&isolated.Daily{},
	)
	return nil
}
