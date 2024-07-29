package tradings

import "time"

type Trigger struct {
  ID           string    `gorm:"size:20;primaryKey"`
  Symbol       string    `gorm:"size:20;not null"`
  TriggerID    string    `gorm:"size:20;index:idx_binance_margin_cross_tradings_triggers"`
  BuyPrice     float64   `gorm:"not null"`
  SellPrice    float64   `gorm:"not null"`
  BuyQuantity  float64   `gorm:"not null"`
  SellQuantity float64   `gorm:"not null"`
  BuyOrderId   int64     `gorm:"not null"`
  SellOrderId  int64     `gorm:"not null"`
  Status       int       `gorm:"not null;index;index:idx_binance_margin_cross_tradings_triggers"`
  Version      int       `gorm:"not null"`
  Remark       string    `gorm:"size:5000;not null"`
  CreatedAt    time.Time `gorm:"not null"`
  UpdatedAt    time.Time `gorm:"not null"`
}

func (m *Trigger) TableName() string {
  return "binance_margin_cross_tradings_triggers"
}
