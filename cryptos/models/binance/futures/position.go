package futures

import "time"

type Position struct {
  ID         string    `gorm:"size:20;primaryKey"`
  Symbol     string    `gorm:"size:20;not null;index:idx_binance_futures_positions_symbol_status"`
  Leverage   int       `gorm:"not null"`
  Side       int       `gorm:"not null"`
  EntryPrice float64   `gorm:"not null"`
  Volume     float64   `gorm:"not null"`
  Notional   float64   `gorm:"not null"`
  Timestamp  int64     `gorm:"not null"`
  Status     int       `gorm:"not null;index:idx_binance_futures_positions_symbol_status"`
  CreatedAt  time.Time `gorm:"not null;index"`
  UpdatedAt  time.Time `gorm:"not null;index"`
}

func (m *Position) TableName() string {
  return "binance_futures_positions"
}
