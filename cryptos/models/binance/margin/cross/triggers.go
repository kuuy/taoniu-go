package cross

import "time"

type Trigger struct {
  ID          string    `gorm:"size:20;primaryKey"`
  Symbol      string    `gorm:"size:20;not null;index:idx_binance_margin_cross_triggers_symbol_status"`
  Side        int       `gorm:"not null"`
  Capital     float64   `gorm:"not null"`
  Price       float64   `gorm:"not null"`
  TakePrice   float64   `gorm:"not null"`
  StopPrice   float64   `gorm:"not null"`
  TakeOrderId int64     `gorm:"not null"`
  StopOrderId int64     `gorm:"not null"`
  Profit      float64   `gorm:"not null"`
  Timestamp   int64     `gorm:"not null"`
  Status      int       `gorm:"size:30;not null;index;index:idx_binance_margin_cross_triggers_symbol_status"`
  Version     int       `gorm:"not null"`
  Remark      string    `gorm:"size:5000;not null"`
  ExpiredAt   time.Time `gorm:"not null"`
  CreatedAt   time.Time `gorm:"not null"`
  UpdatedAt   time.Time `gorm:"not null"`
}

func (m *Trigger) TableName() string {
  return "binance_margin_cross_triggers"
}
