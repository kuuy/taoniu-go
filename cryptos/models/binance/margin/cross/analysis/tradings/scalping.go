package tradings

import (
  "gorm.io/datatypes"
  "time"
)

type Scalping struct {
  ID             string         `gorm:"size:20;primaryKey"`
  Side           int            `gorm:"not null;uniqueIndex:unq_binance_margin_cross_analysis_tradings_scalping"`
  Day            datatypes.Date `gorm:"not null;uniqueIndex:unq_binance_margin_cross_analysis_tradings_scalping"`
  BuysCount      int            `gorm:"not null"`
  SellsCount     int            `gorm:"not null"`
  BuysAmount     float64        `gorm:"type:double precision;not null"`
  SellsAmount    float64        `gorm:"type:double precision;not null"`
  Profit         float64        `gorm:"type:double precision;not null"`
  AdditiveProfit float64        `gorm:"type:double precision;not null"`
  CreatedAt      time.Time      `gorm:"not null"`
  UpdatedAt      time.Time      `gorm:"not null;index"`
}

func (m *Scalping) TableName() string {
  return "binance_margin_cross_analysis_tradings_scalping"
}
