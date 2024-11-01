package tradings

import (
  "gorm.io/datatypes"
  "time"
)

type Scalping struct {
  ID             string         `gorm:"size:20;primaryKey"`
  Day            datatypes.Date `gorm:"not null;uniqueIndex"`
  BuysCount      int            `gorm:"not null"`
  SellsCount     int            `gorm:"not null"`
  BuysAmount     float64        `gorm:"not null"`
  SellsAmount    float64        `gorm:"not null"`
  Profit         float64        `gorm:"not null"`
  AdditiveProfit float64        `gorm:"not null"`
  CreatedAt      time.Time      `gorm:"not null"`
  UpdatedAt      time.Time      `gorm:"not null;index"`
}

func (m *Scalping) TableName() string {
  return "binance_spot_analysis_tradings_scalping"
}
