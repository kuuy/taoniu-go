package futures

import (
  "time"
)

type Order struct {
  ID               string    `gorm:"size:20;primaryKey"`
  Symbol           string    `gorm:"size:20;not null;uniqueIndex:unq_binance_futures_orders;index:idx_binance_futures_orders,priority:1;index:idx_binance_futures_orders_lost"`
  OrderId          int64     `gorm:"not null;uniqueIndex:unq_binance_futures_orders"`
  Type             string    `gorm:"size:30;not null"`
  PositionSide     string    `gorm:"size:20;not null;index:idx_binance_futures_orders,priority:2;index:idx_binance_futures_orders_lost"`
  Side             string    `gorm:"size:20;not null;index:idx_binance_futures_orders_lost"`
  Price            float64   `gorm:"type:double precision;not null"`
  AvgPrice         float64   `gorm:"type:double precision;not null"`
  ActivatePrice    float64   `gorm:"type:double precision;not null"`
  StopPrice        float64   `gorm:"type:double precision;not null"`
  PriceRate        float64   `gorm:"type:double precision;not null"`
  Quantity         float64   `gorm:"type:double precision;not null;index:idx_binance_futures_orders_lost"`
  ExecutedQuantity float64   `gorm:"type:double precision;not null"`
  OpenTime         int64     `gorm:"not null"`
  UpdateTime       int64     `gorm:"not null;index:idx_binance_futures_orders,priority:4;index:idx_binance_futures_orders_lost"`
  WorkingType      string    `gorm:"size:30;not null"`
  PriceProtect     bool      `gorm:"not null"`
  ReduceOnly       bool      `gorm:"not null"`
  ClosePosition    bool      `gorm:"not null"`
  Status           string    `gorm:"size:30;not null;index;index:idx_binance_futures_orders,priority:3"`
  Remark           string    `gorm:"size:5000;not null"`
  CreatedAt        time.Time `gorm:"not null"`
  UpdatedAt        time.Time `gorm:"not null"`
}

func (m *Order) TableName() string {
  return "binance_futures_orders"
}
