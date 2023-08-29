package dydx

import "time"

type Position struct {
  ID            string    `gorm:"size:20;primaryKey"`
  Symbol        string    `gorm:"size:20;not null;index:idx_dydx_positions_symbol_status"`
  Side          int       `gorm:"not null"`
  Leverage      int       `gorm:"not null"`
  Capital       float64   `gorm:"not null"`
  Notional      float64   `gorm:"not null"`
  EntryPrice    float64   `gorm:"not null"`
  EntryQuantity float64   `gorm:"not null"`
  Timestamp     int64     `gorm:"not null"`
  Status        int       `gorm:"not null;index:idx_dydx_positions_symbol_status"`
  Version       int       `gorm:"not null"`
  CreatedAt     time.Time `gorm:"not null;index"`
  UpdatedAt     time.Time `gorm:"not null;index"`
}

func (m *Position) TableName() string {
  return "dydx_positions"
}
