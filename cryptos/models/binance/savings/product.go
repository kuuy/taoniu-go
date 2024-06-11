package savings

import (
  "time"
)

type FlexibleProduct struct {
  ID                       string    `gorm:"size:20;primaryKey"`
  Asset                    string    `gorm:"size:20;not null;uniqueIndex"`
  ProductId                string    `gorm:"size:30;not null"`
  AvgAnnualInterestRate    float64   `gorm:"not null"`
  DailyInterestPerThousand float64   `gorm:"not null"`
  MinPurchaseAmount        float64   `gorm:"not null"`
  PurchasedAmount          float64   `gorm:"not null"`
  UpLimit                  float64   `gorm:"not null"`
  UpLimitPerUser           float64   `gorm:"not null"`
  CanPurchase              bool      `gorm:"not null;"`
  CanRedeem                bool      `gorm:"not null;"`
  Featured                 bool      `gorm:"not null;"`
  Status                   string    `gorm:"not null;size:20;index"`
  CreatedAt                time.Time `gorm:"not null"`
  UpdatedAt                time.Time `gorm:"not null"`
}

func (m *FlexibleProduct) TableName() string {
  return "binance_savings_flexiable_products"
}
