package isolated

import "time"

type TradingGrid struct {
	ID           string    `gorm:"size:20;primaryKey"`
	Symbol       string    `gorm:"size:20;not null"`
	GridID       string    `gorm:"size:20;index;index:idx_binance_spot_margin_isolated_tradings_grids_grid_status"`
	BuyOrderId   int64     `gorm:"not null"`
	SellOrderId  int64     `gorm:"not null"`
	BuyPrice     float64   `gorm:"not null"`
	SellPrice    float64   `gorm:"not null"`
	BuyQuantity  float64   `gorm:"not null"`
	SellQuantity float64   `gorm:"not null"`
	Status       int       `gorm:"size:30;not null;index:idx_binance_spot_margin_isolated_tradings_grids_grid_status"`
	Remark       string    `gorm:"size:5000;not null"`
	CreatedAt    time.Time `gorm:"not null;index"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (m *TradingGrid) TableName() string {
	return "binance_spot_margin_isolated_tradings_grids"
}
