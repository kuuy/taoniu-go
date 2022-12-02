package models

import (
	"gorm.io/gorm"
	"taoniu.local/security/models/gfw"
)

type Gfw struct{}

func NewGfw() *Gfw {
	return &Gfw{}
}

func (m *Gfw) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&gfw.Dns{},
	)
	return nil
}
