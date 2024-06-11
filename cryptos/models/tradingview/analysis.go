package tradingview

import (
  "gorm.io/datatypes"
  "time"
)

type Analysis struct {
  ID             string            `gorm:"size:20;primaryKey"`
  Exchange       string            `gorm:"size:20;not null;uniqueIndex:unq_tradingview_cryptos_analysis_exchange_symbol_interval"`
  Symbol         string            `gorm:"size:20;not null;uniqueIndex:unq_tradingview_cryptos_analysis_exchange_symbol_interval"`
  Interval       string            `gorm:"size:20;not null;uniqueIndex:unq_tradingview_cryptos_analysis_exchange_symbol_interval"`
  Oscillators    datatypes.JSONMap `gorm:"not null"`
  MovingAverages datatypes.JSONMap `gorm:"not null"`
  Indicators     datatypes.JSONMap `gorm:"not null"`
  Summary        datatypes.JSONMap `gorm:"not null"`
  CreatedAt      time.Time         `gorm:"not null"`
  UpdatedAt      time.Time         `gorm:"not null"`
}

func (m *Analysis) TableName() string {
  return "tradingview_cryptos_analysis"
}
