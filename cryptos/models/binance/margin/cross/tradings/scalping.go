package tradings

import "time"

type Scalping struct {
  ID           string    `gorm:"size:20;primaryKey"`
  Symbol       string    `gorm:"size:20;not null"`
  ScalpingId   string    `gorm:"size:20;index:idx_binance_margin_cross_tradings_scalping"`
  PlanId       string    `gorm:"size:20"`
  BuyPrice     float64   `gorm:"type:double precision;not null"`
  SellPrice    float64   `gorm:"type:double precision;not null"`
  BuyQuantity  float64   `gorm:"type:double precision;not null"`
  SellQuantity float64   `gorm:"type:double precision;not null"`
  BuyOrderId   int64     `gorm:"not null"`
  SellOrderId  int64     `gorm:"not null"`
  Status       int       `gorm:"not null;index;index:idx_binance_margin_cross_tradings_scalping"`
  Version      int       `gorm:"not null"`
  Remark       string    `gorm:"size:5000;not null"`
  CreatedAt    time.Time `gorm:"not null"`
  UpdatedAt    time.Time `gorm:"not null"`
}

func (m *Scalping) TableName() string {
  return "binance_margin_cross_tradings_scalping"
}
