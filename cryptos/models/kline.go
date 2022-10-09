package models

import (
	"time"
)

type Kline5s struct {
	ID        string    `gorm:"size:20;primaryKey"`
	Symbol    string    `gorm:"size:20;not null;uniqueIndex:kline5s_symbol_timestamp_key"`
	Price     float64   `gorm:"not null"`
	High      float64   `gorm:"not null"`
	Low       float64   `gorm:"not null"`
	Volume    float64   `gorm:"not null"`
	Quota     float64   `gorm:"not null"`
	Timestamp int64     `gorm:"not null;uniqueIndex:kline5s_symbol_timestamp_key"`
	CreatedAt time.Time `gorm:"not null;index"`
	UpdatedAt time.Time `gorm:"not null"`
}

type Kline1d struct {
	ID        string    `gorm:"size:20;primaryKey"`
	Symbol    string    `gorm:"size:20;not null;uniqueIndex:kline1d_symbol_timestamp_key"`
	Price     float64   `gorm:"not null"`
	High      float64   `gorm:"not null"`
	Low       float64   `gorm:"not null"`
	Volume    float64   `gorm:"not null"`
	Quota     float64   `gorm:"not null"`
	Timestamp int64     `gorm:"not null;uniqueIndex:kline1d_symbol_timestamp_key"`
	CreatedAt time.Time `gorm:"not null;index"`
	UpdatedAt time.Time `gorm:"not null"`
}
