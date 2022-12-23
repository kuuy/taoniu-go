package models

import (
	"gorm.io/gorm"
	"taoniu.local/security/models/tor"
)

type Tor struct{}

func NewTor() *Tor {
	return &Tor{}
}

func (m *Tor) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&tor.Bridge{},
	)
	return nil
}
