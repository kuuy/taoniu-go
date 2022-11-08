package models

import (
	"gorm.io/gorm"
	"taoniu.local/gamblings/models/wolf"
)

type Wolf struct{}

func NewWolf() *Wolf {
	return &Wolf{}
}

func (m *Wolf) AutoMigrate(db *gorm.DB) error {
	wolf.NewDice().AutoMigrate(db)
	return nil
}
