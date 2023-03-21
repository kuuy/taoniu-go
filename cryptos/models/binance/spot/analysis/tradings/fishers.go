package tradings

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/binance/spot/analysis/tradings/fishers"
)

type Fishers struct{}

func NewFishers() *Fishers {
	return &Fishers{}
}

func (m *Fishers) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&fishers.Grid{},
	)
	return nil
}
