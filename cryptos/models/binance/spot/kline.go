package spot

import (
	"time"
)

type Kline1d struct {
	ID        string    `gorm:"size:20;primaryKey"`
	Symbol    string    `gorm:"size:20;not null;uniqueIndex:unq_binance_spot_klines_1d_symbol_timestamp"`
	Open      float64   `gorm:"not null"`
	Close     float64   `gorm:"not null"`
	High      float64   `gorm:"not null"`
	Low       float64   `gorm:"not null"`
	Volume    float64   `gorm:"not null;"`
	Quota     float64   `gorm:"not null"`
	Timestamp int64     `gorm:"not null;uniqueIndex:unq_binance_spot_klines_1d_symbol_timestamp"`
	CreatedAt time.Time `gorm:"not null;index"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (m *Kline1d) TableName() string {
	return "binance_spot_klines_1d"
}
