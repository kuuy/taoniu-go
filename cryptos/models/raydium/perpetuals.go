package raydium

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/raydium/perpetuals"
)

type Perpetuals struct{}

func NewPerpetuals() *Perpetuals {
	return &Perpetuals{}
}

func (m *Perpetuals) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&perpetuals.Kline{},
	)

	return nil
}
