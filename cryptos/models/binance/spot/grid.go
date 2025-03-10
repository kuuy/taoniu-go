package spot

import (
  "time"
)

type Grid struct {
  ID                string  `gorm:"size:20;primaryKey"`
  Symbol            string  `gorm:"size:20;not null;index:idx_binance_spot_grids_symbol_status"`
  Step              int64   `gorm:"not null;index:idx_binance_spot_grids_symbol_status"`
  Balance           float64 `gorm:"type:double precision;not null"`
  Amount            float64 `gorm:"type:double precision;not null"`
  Quantity          float64 `gorm:"type:double precision;not null"`
  ProfitTarget      float64 `gorm:"type:double precision;not null"`
  StopLossPoint     float64 `gorm:"type:double precision;not null"`
  TakeProfitPrice   float64 `gorm:"type:double precision;not null"`
  TriggerPercent    float64 `gorm:"type:double precision;not null"`
  TakeProfitPercent float64 `gorm:"type:double precision;not null"`
  Status            int64   `gorm:"not null;index:idx_binance_spot_grids_symbol_status"`
  Remark            string  `gorm:"size:5000;not null"`
  CreatedAt         time.Time `gorm:"not null"`
  UpdatedAt         time.Time `gorm:"not null"`
}

func (m *Grid) TableName() string {
  return "binance_spot_grids"
}
