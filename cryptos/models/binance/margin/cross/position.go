package cross

import "time"

type Position struct {
  ID            string    `gorm:"size:20;primaryKey"`
  Symbol        string    `gorm:"size:20;not null;index:idx_binance_margin_cross_positions"`
  Side          int       `gorm:"type:integer;not null"`
  Leverage      int       `gorm:"type:integer;not null"`
  Capital       float64   `gorm:"type:double precision;not null"`
  Notional      float64   `gorm:"type:double precision;not null"`
  EntryPrice    float64   `gorm:"type:double precision;not null"`
  EntryQuantity float64   `gorm:"type:double precision;not null"`
  EntryAmount   float64   `gorm:"type:double precision;not null"`
  Timestamp     int64     `gorm:"not null"`
  Status        int       `gorm:"type:integer;not null;index:idx_binance_margin_cross_positions"`
  Version       int       `gorm:"not null"`
  CreatedAt     time.Time `gorm:"not null;index"`
  UpdatedAt     time.Time `gorm:"not null;index"`
}

func (m *Position) TableName() string {
  return "binance_margin_cross_positions"
}
