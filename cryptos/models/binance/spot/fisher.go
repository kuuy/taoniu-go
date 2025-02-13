package spot

import (
  "gorm.io/datatypes"
  "time"
)

type Fisher struct {
  ID            string         `gorm:"size:20;primaryKey"`
  Symbol        string         `gorm:"size:20;not null"`
  Price         float64        `gorm:"type:double precision;not null"`
  Balance       float64        `gorm:"type:double precision;not null"`
  Tickers       datatypes.JSON `gorm:"not null"`
  StartAmount   float64        `gorm:"type:double precision;not null"`
  StartBalance  float64        `gorm:"type:double precision;not null"`
  TargetBalance float64        `gorm:"type:double precision;not null"`
  StopBalance   float64        `gorm:"type:double precision;not null"`
  Status        int            `gorm:"not null"`
  Remark        string         `gorm:"size:5000;not null"`
  CreatedAt     time.Time      `gorm:"not null"`
  UpdatedAt     time.Time      `gorm:"not null"`
}

func (m *Fisher) TableName() string {
  return "binance_spot_fishers"
}
