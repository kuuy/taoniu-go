package isolated

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/binance/spot/margin/isolated/fishers"
)

type Fishers struct{}

func NewFishers() *Fishers {
	return &Fishers{}
}

func (m *Fishers) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&fishers.Fisher{},
		&fishers.Grid{},
	)
	return nil
}
