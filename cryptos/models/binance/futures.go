package binance

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/binance/futures"
)

type Futures struct{}

func NewFutures() *Futures {
	return &Futures{}
}

func (m *Futures) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&futures.Symbol{},
		&futures.Order{},
		&futures.Kline{},
		&futures.Strategy{},
		&futures.Plan{},
		&futures.Grid{},
		&futures.TradingScalping{},
		&futures.TradingGrid{},
	)

	return nil
}
