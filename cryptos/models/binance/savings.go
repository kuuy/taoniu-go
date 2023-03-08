package binance

import (
  "gorm.io/gorm"
  "taoniu.local/cryptos/models/binance/savings"
)

type Savings struct{}

func NewSavings() *Savings {
	return &Savings{}
}

func (m *Savings) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&savings.FlexibleProduct{},
	)

	return nil
}
