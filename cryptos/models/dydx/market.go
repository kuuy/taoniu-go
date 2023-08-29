package dydx

import (
  "time"
)

type Market struct {
  ID              string    `gorm:"size:20;primaryKey"`
  Symbol          string    `gorm:"size:20;not null;uniqueIndex"`
  BaseAsset       string    `gorm:"not null"`
  QuoteAsset      string    `gorm:"not null"`
  StepSize        float64   `gorm:"not null"`
  TickSize        float64   `gorm:"not null"`
  MinOrderSize    float64   `gorm:"not null"`
  MaxPositionSize float64   `gorm:"not null"`
  Status          string    `gorm:"not null;size:20;index"`
  CreatedAt       time.Time `gorm:"not null"`
  UpdatedAt       time.Time `gorm:"not null"`
}

func (m *Market) TableName() string {
  return "dydx_markets"
}
