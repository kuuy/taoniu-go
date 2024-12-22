package tradings

import (
  "gorm.io/datatypes"
  "time"
)

type Scalping struct {
  ID             string         `gorm:"size:20;primaryKey"`
  Day            datatypes.Date `gorm:"not null;uniqueIndex"`
  BuysCount      int            `gorm:"type:integer;not null"`
  SellsCount     int            `gorm:"type:integer;not null"`
  BuysAmount     float64        `gorm:"type:double precision;not null"`
  SellsAmount    float64        `gorm:"type:double precision;not null"`
  Profit         float64        `gorm:"type:double precision;not null"`
  AdditiveProfit float64        `gorm:"type:double precision;not null"`
  CreatedAt      time.Time      `gorm:"not null"`
  UpdatedAt      time.Time      `gorm:"not null;index"`
}

func (m *Scalping) TableName() string {
  return "binance_spot_analysis_tradings_scalping"
}
