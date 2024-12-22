package models

import "time"

type Markets struct {
  ID         string    `gorm:"size:20;primaryKey"`
  ExchangeID string    `gorm:"size:20;not null;uniqueIndex"`
  CurrencyID string    `gorm:"size:20;not null"`
  Symbol     string    `gorm:"size:20;not null"`
  Price      float64   `gorm:"type:double precision;not null"`
  Volume     float64   `gorm:"type:double precision;not null"`
  Liquidity  float64   `gorm:"not null;index"`
  Status     int       `gorm:"type:integer;not null;index"`
  About      string    `gorm:"size:20000;not null"`
  CreatedAt  time.Time `gorm:"not null"`
  UpdatedAt  time.Time `gorm:"not null"`
}

func (m *Markets) TableName() string {
  return "cryptos_markets"
}
