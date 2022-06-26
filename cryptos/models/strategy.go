package models

import (
	"time"
)

type Strategy struct {
  ID string `gorm:"size:20;primaryKey"`
  Symbol string `gorm:"size:20;not null;uniqueIndex:strategy_symbol_indicator_timestamp_key"`
  Indicator string `gorm:"size:30;not null;uniqueIndex:strategy_symbol_indicator_timestamp_key"`
  Price float64 `gorm:"not null"`
  Signal int64 `gorm:"not null"`
  Volume float64 `gorm:"not null"`
  Timestamp int64 `gorm:"not null;uniqueIndex:strategy_symbol_indicator_timestamp_key"`
  Remark string `gorm:"size:5000;not null"`
  CreatedAt time.Time `gorm:"not null;index"`
  UpdatedAt time.Time `gorm:"not null;index"`
}

