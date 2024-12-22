package models

import (
  "gorm.io/datatypes"
  "time"
)

type Currency struct {
  ID                string         `gorm:"size:20;primaryKey"`
  Symbol            string         `gorm:"size:20;not null;uniqueIndex"`
  Type              int            `gorm:"type:integer;not null"`
  SectorId          string         `gorm:"size:20;not null;index"`
  TotalSupply       float64        `gorm:"type:double precision;not null"`
  CirculatingSupply float64        `gorm:"type:double precision;not null"`
  Price             float64        `gorm:"type:double precision;not null"`
  Volume            float64        `gorm:"type:double precision;not null"`
  MarketCap         float64        `gorm:"type:double precision;not null"`
  Exchanges         datatypes.JSON `gorm:"size:2000;not null"`
  Status            int            `gorm:"type:integer;not null;index"`
  About             string         `gorm:"size:20000;not null"`
  CreatedAt         time.Time      `gorm:"not null"`
  UpdatedAt         time.Time      `gorm:"not null"`
}

func (m *Currency) TableName() string {
  return "cryptos_currencies"
}
