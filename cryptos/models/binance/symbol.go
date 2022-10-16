package binance

import (
	"gorm.io/datatypes"
	"time"
)

type Symbol struct {
	ID         string            `gorm:"size:20;primaryKey"`
	Symbol     string            `gorm:"size:20;not null;uniqueIndex"`
	BaseAsset  string            `gorm:"not null"`
	QuoteAsset string            `gorm:"not null"`
	Filters    datatypes.JSONMap `gorm:"not null"`
	IsSpot     bool              `gorm:"not null;"`
	IsMargin   bool              `gorm:"not null;"`
	IsFutures  bool              `gorm:"not null;"`
	Status     string            `gorm:"not null;size:20;index"`
	CreatedAt  time.Time         `gorm:"not null"`
	UpdatedAt  time.Time         `gorm:"not null"`
}

func (m *Symbol) TableName() string {
	return "binance_symbols"
}
