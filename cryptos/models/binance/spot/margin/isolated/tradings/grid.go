package tradings

import "time"

type Grid struct {
	ID           string    `gorm:"size:20;primaryKey"`
	Symbol       string    `gorm:"size:20;not null;index:idx_binance_spot_margin_isolated_tradings_grids_created_symbol,priority:2"`
	GridID       string    `gorm:"size:20;index;index:idx_binance_spot_margin_isolated_tradings_grids_grid_status"`
	BuyOrderId   int64     `gorm:"not null"`
	SellOrderId  int64     `gorm:"not null"`
	BuyPrice     float64   `gorm:"not null"`
	SellPrice    float64   `gorm:"not null"`
	BuyQuantity  float64   `gorm:"not null"`
	SellQuantity float64   `gorm:"not null"`
	Status       int       `gorm:"not null;index:idx_binance_spot_margin_isolated_tradings_grids_grid_status"`
	Remark       string    `gorm:"size:5000;not null"`
	CreatedAt    time.Time `gorm:"not null;index;index:idx_binance_spot_margin_isolated_tradings_grids_created_symbol,priority:1"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (m *Grid) TableName() string {
	return "binance_spot_margin_isolated_tradings_grids"
}
