package models

import (
	"time"
)

type User struct {
	ID        string    `gorm:"size:20;primaryKey"`
	Email     string    `gorm:"size:64;not null;uniqueIndex"`
	Password  string    `gorm:"size:128;not null"`
	Salt      string    `gorm:"size:16;not null"`
	Status    int64     `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (m *User) TableName() string {
	return "users"
}
