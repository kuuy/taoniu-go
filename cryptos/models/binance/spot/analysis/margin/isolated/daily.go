package isolated

import (
	"gorm.io/datatypes"
	"time"
)

type Daily struct {
	ID          string         `gorm:"size:20;primaryKey"`
	Symbol      string         `gorm:"size:20;not null;uniqueIndex:unq_binance_spot_analysis_margin_isolated_daily_symbol_day"`
	Day         datatypes.Date `gorm:"not null;uniqueIndex:unq_binance_spot_analysis_margin_isolated_daily_symbol_day"`
	TotalProfit float64        `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
}

func (m *Daily) TableName() string {
	return "binance_spot_analysis_margin_isolated_daily"
}
