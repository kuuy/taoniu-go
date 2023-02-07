package isolated

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/binance/spot/margin/isolated/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
	return &Tradings{}
}

func (m *Tradings) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&tradings.Grid{},
	)
	tradings.NewFishers().AutoMigrate(db)
	return nil
}
