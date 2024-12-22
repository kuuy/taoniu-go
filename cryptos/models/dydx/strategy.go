package dydx

import "time"

type Strategy struct {
  ID        string    `gorm:"size:20;primaryKey"`
  Symbol    string    `gorm:"size:20;not null;uniqueIndex:unq_dydx_strategies"`
  Indicator string    `gorm:"size:30;not null;uniqueIndex:unq_dydx_strategies"`
  Interval  string    `gorm:"size:3;not null;uniqueIndex:unq_dydx_strategies"`
  Price     float64   `gorm:"type:double precision;not null"`
  Signal    int       `gorm:"type:integer;not null"`
  Timestamp int64     `gorm:"not null;uniqueIndex:unq_dydx_strategies"`
  CreatedAt time.Time `gorm:"not null;index"`
  UpdatedAt time.Time `gorm:"not null;index"`
}

func (m *Strategy) TableName() string {
  return "dydx_strategies"
}
