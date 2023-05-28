package fishers

import (
	"gorm.io/datatypes"
	"time"
)

type Grid struct {
	ID          string            `gorm:"size:20;primaryKey"`
	Day         datatypes.Date    `gorm:"not null;uniqueIndex"`
	BuysCount   int               `gorm:"not null"`
	SellsCount  int               `gorm:"not null"`
	BuysAmount  float64           `gorm:"not null"`
	SellsAmount float64           `gorm:"not null"`
	Profit      float64           `gorm:"not null"`
	Data        datatypes.JSONMap `gorm:"not null"`
	CreatedAt   time.Time         `gorm:"not null"`
	UpdatedAt   time.Time         `gorm:"not null;index"`
}

func (m *Grid) TableName() string {
	return "binance_spot_analysis_margin_isolated_tradings_fishers_grids"
}