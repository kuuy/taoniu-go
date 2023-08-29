package tradings

import "time"

type Scalping struct {
  ID           string    `gorm:"size:20;primaryKey"`
  Symbol       string    `gorm:"size:20;not null"`
  ScalpingID   string    `gorm:"size:20;index"`
  PlanID       string    `gorm:"size:20;index"`
  BuyPrice     float64   `gorm:"not null"`
  SellPrice    float64   `gorm:"not null"`
  BuyQuantity  float64   `gorm:"not null"`
  SellQuantity float64   `gorm:"not null"`
  BuyOrderId   string    `gorm:"size:63;not null"`
  SellOrderId  string    `gorm:"size:63;not null"`
  Status       int       `gorm:"not null;index:idx_dydx_tradings_scalping_updated_status,priority:2"`
  Version      int       `gorm:"not null"`
  Remark       string    `gorm:"size:5000;not null"`
  CreatedAt    time.Time `gorm:"not null"`
  UpdatedAt    time.Time `gorm:"not null;index;index:idx_dydx_tradings_scalping_updated_status,priority:1"`
}

func (m *Scalping) TableName() string {
  return "dydx_tradings_scalping"
}
