package margin

import (
	"gorm.io/datatypes"
	"time"
)

type Isolated struct {
	ID               string         `gorm:"size:20;primaryKey"`
	Day              datatypes.Date `gorm:"not null;uniqueIndex"`
	BuysCount        float64        `gorm:"not null"`
	SellsCount       float64        `gorm:"not null"`
	GridsBuysCount   float64        `gorm:"not null"`
	GridsSellsCount  float64        `gorm:"not null"`
	GridsBuysAmount  float64        `gorm:"not null"`
	GridsSellsAmount float64        `gorm:"not null"`
	GridsProfit      float64        `gorm:"not null"`
	TotalProfit      float64        `gorm:"not null"`
	CreatedAt        time.Time      `gorm:"not null"`
	UpdatedAt        time.Time      `gorm:"not null;index"`
}

func (m *Isolated) TableName() string {
	return "binance_spot_analysis_daily_margin_isolated"
}
