package binance

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/binance/spot"
)

type Spot struct{}

func NewSpot() *Spot {
	return &Spot{}
}

func (m *Spot) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&spot.Order{},
		&spot.Kline1d{},
		&spot.Strategy{},
	)
	spot.NewMargin().AutoMigrate(db)
	spot.NewAnalysis().AutoMigrate(db)
	return nil
}
