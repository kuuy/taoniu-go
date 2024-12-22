package tradings

import (
  "gorm.io/datatypes"
  "time"
)

type Trigger struct {
  ID             string         `gorm:"size:20;primaryKey"`
  Side           int            `gorm:"type:integer;not null;uniqueIndex:unq_binance_futures_analysis_tradings_trigger"`
  Day            datatypes.Date `gorm:"not null;uniqueIndex:unq_binance_futures_analysis_tradings_trigger"`
  BuysCount      int            `gorm:"type:integer;not null"`
  SellsCount     int            `gorm:"type:integer;not null"`
  BuysAmount     float64        `gorm:"type:double precision;not null"`
  SellsAmount    float64        `gorm:"type:double precision;not null"`
  Profit         float64        `gorm:"type:double precision;not null"`
  AdditiveProfit float64        `gorm:"type:double precision;not null"`
  CreatedAt      time.Time      `gorm:"not null"`
  UpdatedAt      time.Time      `gorm:"not null;index"`
}

func (m *Trigger) TableName() string {
  return "binance_futures_analysis_tradings_triggers"
}
