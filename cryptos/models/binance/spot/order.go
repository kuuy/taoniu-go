package spot

import "time"

type Order struct {
  ID               string    `gorm:"size:20;primaryKey"`
  Symbol           string    `gorm:"size:20;not null;uniqueIndex:unq_binance_spot_orders;index:idx_binance_spot_orders,priority:1;index:idx_binance_spot_orders_lost"`
  OrderId          int64     `gorm:"not null;uniqueIndex:unq_binance_spot_orders"`
  Type             string    `gorm:"size:30;not null"`
  Side             string    `gorm:"size:20;not null;index:idx_binance_spot_orders_lost"`
  Price            float64   `gorm:"not null"`
  AvgPrice         float64   `gorm:"not null"`
  StopPrice        float64   `gorm:"not null"`
  Quantity         float64   `gorm:"not null;index:idx_binance_spot_orders_lost"`
  ExecutedQuantity float64   `gorm:"not null"`
  OpenTime         int64     `gorm:"not null;"`
  UpdateTime       int64     `gorm:"not null;index:idx_binance_spot_orders,priority:3;index:idx_binance_spot_orders_lost"`
  Status           string    `gorm:"size:30;not null;index;index:idx_binance_spot_orders,priority:2"`
  Remark           string    `gorm:"size:5000;not null"`
  CreatedAt        time.Time `gorm:"not null"`
  UpdatedAt        time.Time `gorm:"not null"`
}

func (m *Order) TableName() string {
  return "binance_spot_orders"
}
