package models

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/binance"
)

type Binance struct{}

func NewBinance() *Binance {
	return &Binance{}
}

func (m *Binance) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&binance.Symbol{},
	)
	binance.NewSpot().AutoMigrate(db)
	return nil
}
