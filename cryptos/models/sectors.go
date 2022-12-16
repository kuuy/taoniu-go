package models

import "time"

type Sector struct {
	ID        string    `gorm:"size:20;primaryKey"`
	Name      string    `gorm:"size:50;not null"`
	Slug      string    `gorm:"size:50;not null;uniqueIndex"`
	Status    int       `gorm:"not null;index"`
	About     string    `gorm:"size:20000;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (m *Sector) TableName() string {
	return "cryptos_sectors"
}
