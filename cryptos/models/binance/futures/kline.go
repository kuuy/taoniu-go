package futures

import "time"

type Kline struct {
  ID        string    `gorm:"size:20;primaryKey"`
  Symbol    string    `gorm:"size:20;not null;uniqueIndex:unq_binance_futures_klines"`
  Interval  string    `gorm:"size:3;not null;uniqueIndex:unq_binance_futures_klines"`
  Open      float64   `gorm:"type:double precision;not null"`
  Close     float64   `gorm:"type:double precision;not null"`
  High      float64   `gorm:"type:double precision;not null"`
  Low       float64   `gorm:"type:double precision;not null"`
  Volume    float64   `gorm:"type:double precision;not null;"`
  Quota     float64   `gorm:"type:double precision;not null"`
  Timestamp int64     `gorm:"not null;uniqueIndex:unq_binance_futures_klines"`
  CreatedAt time.Time `gorm:"not null"`
  UpdatedAt time.Time `gorm:"not null"`
}

func (m *Kline) TableName() string {
  return "binance_futures_klines"
}
