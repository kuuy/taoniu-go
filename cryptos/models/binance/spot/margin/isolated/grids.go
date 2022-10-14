package isolated

import "time"

type Grids struct {
	ID              string    `gorm:"size:20;primaryKey"`
	Symbol          string    `gorm:"size:20;not null;index:idx_binance_spot_margin_isolated_grids_symbol_status"`
	Step            int64     `gorm:"not null"`
	R3              float64   `gorm:"not null"`
	R2              float64   `gorm:"not null"`
	R1              float64   `gorm:"not null"`
	S1              float64   `gorm:"not null"`
	S2              float64   `gorm:"not null"`
	S3              float64   `gorm:"not null"`
	ProfitTarget    float64   `gorm:"not null"`
	StopLossPoint   float64   `gorm:"not null"`
	TakeProfitPrice float64   `gorm:"not null"`
	TriggerPercent  float64   `gorm:"not null"`
	Balance         float64   `gorm:"not null"`
	OrderId         int64     `gorm:"not null"`
	Status          int64     `gorm:"size:30;not null;index:idx_binance_spot_margin_isolated_grids_symbol_status"`
	Remark          string    `gorm:"size:5000;not null"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

func (m *Grids) TableName() string {
	return "binance_spot_margin_isolated_grids"
}
