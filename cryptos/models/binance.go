package models

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/binance"
	"taoniu.local/cryptos/models/binance/spot"
)

type Binance struct{}

func NewBinance() *Binance {
	return &Binance{}
}

func (m *Binance) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&spot.Symbol{},
	)
	binance.NewSpot().AutoMigrate(db)
	binance.NewFutures().AutoMigrate(db)
	binance.NewSavings().AutoMigrate(db)
	return nil
}
