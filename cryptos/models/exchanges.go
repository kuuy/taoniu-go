package models

import "time"

type Exchange struct {
	ID        string    `gorm:"size:20;primaryKey"`
	Name      string    `gorm:"size:50;not null;uniqueIndex"`
	Volume    float64   `gorm:"not null"`
	Status    int       `gorm:"not null;index"`
	About     string    `gorm:"size:20000;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (m *Exchange) TableName() string {
	return "cryptos_exchanges"
}
