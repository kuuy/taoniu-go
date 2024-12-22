package spot

import "time"

type Position struct {
  ID            string    `gorm:"size:20;primaryKey"`
  Symbol        string    `gorm:"size:20;not null;index:idx_binance_spot_positions_symbol_status"`
  Notional      float64   `gorm:"type:double precision;not null"`
  EntryPrice    float64   `gorm:"type:double precision;not null"`
  EntryQuantity float64   `gorm:"type:double precision;not null"`
  EntryAmount   float64   `gorm:"type:double precision;not null"`
  Timestamp     int64     `gorm:"not null"`
  Status        int       `gorm:"type:integer;not null;index:idx_binance_spot_positions_symbol_status"`
  Version       int       `gorm:"not null"`
  CreatedAt     time.Time `gorm:"not null;index"`
  UpdatedAt     time.Time `gorm:"not null;index"`
}

func (m *Position) TableName() string {
  return "binance_spot_positions"
}
