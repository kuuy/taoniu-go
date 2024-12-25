package gambling

import (
  "gorm.io/datatypes"
  "time"
)

type Ant struct {
  ID              string                       `gorm:"size:20;primaryKey"`
  Symbol          string                       `gorm:"size:20;not null;index:idx_binance_spot_gambling_ant"`
  Mode            int                          `gorm:"not null"`
  EntryPrice      float64                      `gorm:"type:double precision;not null"`
  EntryQuantity   float64                      `gorm:"type:double precision;not null"`
  PlacePrices     datatypes.JSONSlice[float64] `gorm:"not null"`
  PlaceQuantities datatypes.JSONSlice[float64] `gorm:"not null"`
  TakePrices      datatypes.JSONSlice[float64] `gorm:"not null"`
  TakeQuantities  datatypes.JSONSlice[float64] `gorm:"not null"`
  PlaceQuantity   float64                      `gorm:"type:double precision;not null"`
  TakeQuantity    float64                      `gorm:"type:double precision;not null"`
  Profit          float64                      `gorm:"type:double precision;not null"`
  Timestamp       int64                        `gorm:"not null"`
  Status          int                          `gorm:"type:integer;not null;index;index:idx_binance_spot_gambling_ant"`
  Version         int                          `gorm:"not null"`
  Remark          string                       `gorm:"size:5000;not null"`
  ExpiredAt       time.Time                    `gorm:"not null"`
  CreatedAt       time.Time                    `gorm:"not null"`
  UpdatedAt       time.Time                    `gorm:"not null"`
}

func (m *Ant) TableName() string {
  return "binance_spot_gambling_ant"
}
