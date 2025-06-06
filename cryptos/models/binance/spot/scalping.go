package spot

import "time"

type Scalping struct {
  ID          string    `gorm:"size:20;primaryKey"`
  Symbol      string    `gorm:"size:20;not null;index:idx_binance_spot_scalping_symbol_status"`
  Capital     float64   `gorm:"type:double precision;not null"`
  Price       float64   `gorm:"type:double precision;not null"`
  TakePrice   float64   `gorm:"type:double precision;not null"`
  StopPrice   float64   `gorm:"type:double precision;not null"`
  TakeOrderId int64     `gorm:"not null"`
  StopOrderId int64     `gorm:"not null"`
  Profit      float64   `gorm:"type:double precision;not null"`
  Timestamp   int64     `gorm:"not null"`
  Status      int       `gorm:"type:integer;not null;index;index:idx_binance_spot_scalping_symbol_status"`
  Version     int       `gorm:"not null"`
  Remark      string    `gorm:"size:5000;not null"`
  ExpiredAt   time.Time `gorm:"not null"`
  CreatedAt   time.Time `gorm:"not null"`
  UpdatedAt   time.Time `gorm:"not null"`
}

func (m *Scalping) TableName() string {
  return "binance_spot_scalping"
}

type ScalpingPlan struct {
  PlanId string `gorm:"size:20;uniqueIndex"`
  Status int    `gorm:"not null;index"`
}

func (m *ScalpingPlan) TableName() string {
  return "binance_spot_scalping_plans"
}
