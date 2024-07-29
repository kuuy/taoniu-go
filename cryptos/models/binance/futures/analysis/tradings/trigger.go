package tradings

import (
  "gorm.io/datatypes"
  "time"
)

type Trigger struct {
  ID             string         `gorm:"size:20;primaryKey"`
  Side           int            `gorm:"not null;uniqueIndex:unq_binance_futures_analysis_tradings_trigger"`
  Day            datatypes.Date `gorm:"not null;uniqueIndex:unq_binance_futures_analysis_tradings_trigger"`
  BuysCount      int            `gorm:"not null"`
  SellsCount     int            `gorm:"not null"`
  BuysAmount     float64        `gorm:"not null"`
  SellsAmount    float64        `gorm:"not null"`
  Profit         float64        `gorm:"not null"`
  AdditiveProfit float64        `gorm:"not null"`
  CreatedAt      time.Time      `gorm:"not null"`
  UpdatedAt      time.Time      `gorm:"not null;index"`
}

func (m *Trigger) TableName() string {
  return "binance_futures_analysis_tradings_triggers"
}
