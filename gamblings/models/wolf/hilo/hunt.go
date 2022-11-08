package hilo

import (
	"time"
)

type Hunt struct {
	ID        string    `gorm:"size:20;primaryKey"`
	Hash      string    `gorm:"size:36;not null"`
	Rules     string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null;index"`
}

func (m *Hunt) TableName() string {
	return "wolf_hilo_hunts"
}
