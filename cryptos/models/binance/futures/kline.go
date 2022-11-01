package futures

import "time"

type Kline struct {
	ID        string    `gorm:"size:20;primaryKey"`
	Symbol    string    `gorm:"size:20;not null;uniqueIndex:unq_binance_futures_klines_symbol_interval_timestamp"`
	Interval  string    `gorm:"size:2;not null;uniqueIndex:unq_binance_futures_klines_symbol_interval_timestamp"`
	Open      float64   `gorm:"not null"`
	Close     float64   `gorm:"not null"`
	High      float64   `gorm:"not null"`
	Low       float64   `gorm:"not null"`
	Volume    float64   `gorm:"not null;"`
	Quota     float64   `gorm:"not null"`
	Timestamp int64     `gorm:"not null;uniqueIndex:unq_binance_futures_klines_symbol_interval_timestamp"`
	CreatedAt time.Time `gorm:"not null;index"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (m *Kline) TableName() string {
	return "binance_futures_klines"
}
