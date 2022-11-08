package dice

import (
	"time"
)

type Hunt struct {
	ID         string    `gorm:"size:20;primaryKey"`
	Number     float64   `gorm:"not null;unique"`
	Ipart      uint8     `gorm:"not null;index"`
	Dpart      uint8     `gorm:"not null;index"`
	Hash       string    `gorm:"size:36;not null"`
	Side       uint8     `gorm:"not null"`
	IsMirror   bool      `gorm:"not null;index"`
	IsRepeate  bool      `gorm:"not null;index"`
	IsNeighbor bool      `gorm:"not null;index"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null;index"`
}

func (m *Hunt) TableName() string {
	return "wolf_dice_hunts"
}
