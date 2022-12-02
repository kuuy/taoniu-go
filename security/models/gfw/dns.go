package gfw

import (
	"time"
)

type Dns struct {
	ID        string    `gorm:"size:20;primaryKey"`
	Domain    string    `gorm:"size:255;not null"`
	Hash      string    `gorm:"size:64;not null;index"`
	Ips       string    `gorm:"size:5000;not null"`
	Status    int       `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null;index"`
}

func (m *Dns) TableName() string {
	return "gfw_dns"
}
