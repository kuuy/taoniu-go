package spot

import (
	"gorm.io/datatypes"
	"time"
)

type Plan struct {
	ID        string            `gorm:"size:20;primaryKey"`
	Symbol    string            `gorm:"size:20;not null;uniqueIndex:unq_binance_spot_plans_symbol_timestamp"`
	Side      int64             `gorm:"not null"`
	Price     float64           `gorm:"not null"`
	Quantity  float64           `gorm:"not null"`
	Amount    float64           `gorm:"not null"`
	Timestamp int64             `gorm:"not null;index;uniqueIndex:unq_binance_spot_plans_symbol_timestamp"`
	Context   datatypes.JSONMap `gorm:"not null"`
	Status    int64             `gorm:"not null"`
	Remark    string            `gorm:"size:5000;not null"`
	CreatedAt time.Time         `gorm:"not null;index"`
	UpdatedAt time.Time         `gorm:"not null"`
}

func (m *Plan) TableName() string {
	return "binance_spot_plans"
}