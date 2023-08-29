package spot

import "time"

type Order struct {
  ID               string    `gorm:"size:20;primaryKey"`
  Symbol           string    `gorm:"size:20;not null;index:idx_binance_spot_orders_symbol_update_status,priority:2;uniqueIndex:unq_binance_spot_orders_symbol_orderid"`
  OrderID          int64     `gorm:"not null;uniqueIndex:unq_binance_spot_orders_symbol_orderid"`
  Type             string    `gorm:"size:30;not null"`
  Side             string    `gorm:"size:20;not null"`
  Price            float64   `gorm:"not null"`
  StopPrice        float64   `gorm:"not null"`
  Quantity         float64   `gorm:"not null"`
  ExecutedQuantity float64   `gorm:"not null"`
  OpenTime         int64     `gorm:"not null;"`
  UpdateTime       int64     `gorm:"not null;index:idx_binance_spot_orders_symbol_update_status,priority:1"`
  Status           string    `gorm:"size:30;not null;index:idx_binance_spot_orders_symbol_update_status,priority:3"`
  Remark           string    `gorm:"size:5000;not null"`
  CreatedAt        time.Time `gorm:"not null"`
  UpdatedAt        time.Time `gorm:"not null;index"`
}

func (m *Order) TableName() string {
  return "binance_spot_orders"
}
