package spot

import "time"

type Strategy struct {
	ID        string    `gorm:"size:20;primaryKey"`
	Symbol    string    `gorm:"size:20;not null;uniqueIndex:unq_binance_spot_strategies_symbol_indicator_interval_timestamp"`
	Indicator string    `gorm:"size:30;not null;uniqueIndex:unq_binance_spot_strategies_symbol_indicator_interval_timestamp"`
	Interval  string    `gorm:"size:2;not null;uniqueIndex:unq_binance_spot_strategies_symbol_indicator_interval_timestamp"`
	Price     float64   `gorm:"not null"`
	Signal    int64     `gorm:"not null"`
	Timestamp int64     `gorm:"not null;uniqueIndex:unq_binance_spot_strategies_symbol_indicator_interval_timestamp"`
	Remark    string    `gorm:"size:5000;not null"`
	CreatedAt time.Time `gorm:"not null;index"`
	UpdatedAt time.Time `gorm:"not null;index"`
}

func (m *Strategy) TableName() string {
	return "binance_spot_strategies"
}
