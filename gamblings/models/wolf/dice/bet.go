package dice

import (
	"gorm.io/gorm"
)

type Bet struct{}

func NewBet() *Bet {
	return &Bet{}
}

func (m *Bet) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&Multiple{},
	)
	return nil
}
