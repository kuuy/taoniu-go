package wolf

import (
	"gorm.io/gorm"
	"taoniu.local/gamblings/models/wolf/dice"
)

type Dice struct{}

func NewDice() *Dice {
	return &Dice{}
}

func (m *Dice) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&dice.Hunt{},
		&dice.Multiple{},
		&dice.Plan{},
		&dice.Hell{},
	)
	dice.NewBet().AutoMigrate(db)
	return nil
}
