package dydx

import (
  "time"
)

type Order struct {
  ID               string    `gorm:"size:20;primaryKey"`
  Symbol           string    `gorm:"size:20;not null;index"`
  OrderID          string    `gorm:"size:63;not null;uniqueIndex"`
  Type             string    `gorm:"size:30;not null"`
  Side             string    `gorm:"size:20;not null"`
  Price            float64   `gorm:"not null"`
  AvgPrice         float64   `gorm:"not null"`
  ActivatePrice    float64   `gorm:"not null"`
  StopPrice        float64   `gorm:"not null"`
  PriceRate        float64   `gorm:"not null"`
  Quantity         float64   `gorm:"not null"`
  ExecutedQuantity float64   `gorm:"not null"`
  OpenTime         int64     `gorm:"not null;"`
  UpdateTime       int64     `gorm:"not null;"`
  WorkingType      string    `gorm:"size:30;not null"`
  PriceProtect     bool      `gorm:"not null"`
  ReduceOnly       bool      `gorm:"not null"`
  ClosePosition    bool      `gorm:"not null"`
  CancelReason     string    `gorm:"size:100;not null"`
  Status           string    `gorm:"size:30;not null;index"`
  Remark           string    `gorm:"size:5000;not null"`
  CreatedAt        time.Time `gorm:"not null"`
  UpdatedAt        time.Time `gorm:"not null"`
}

func (m *Order) TableName() string {
  return "dydx_orders"
}
