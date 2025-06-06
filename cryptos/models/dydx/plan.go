package dydx

import (
  "gorm.io/datatypes"
  "time"
)

type Plan struct {
  ID        string            `gorm:"size:20;primaryKey"`
  Symbol    string            `gorm:"size:20;not null;uniqueIndex:unq_dydx_plans_symbol_interval_timestamp"`
  Interval  string            `gorm:"size:3;not null;uniqueIndex:unq_dydx_plans_symbol_interval_timestamp"`
  Side      int               `gorm:"type:integer;not null"`
  Price     float64           `gorm:"not null"`
  Quantity  float64           `gorm:"not null"`
  Amount    float64           `gorm:"not null"`
  Timestamp int64             `gorm:"not null;uniqueIndex:unq_dydx_plans_symbol_interval_timestamp"`
  Context   datatypes.JSONMap `gorm:"not null"`
  Status    int               `gorm:"type:integer;not null;index"`
  Remark    string            `gorm:"size:5000;not null"`
  CreatedAt time.Time         `gorm:"not null;index"`
  UpdatedAt time.Time         `gorm:"not null"`
}

func (m *Plan) TableName() string {
  return "dydx_plans"
}
