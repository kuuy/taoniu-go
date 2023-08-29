package spot

import "time"

type Launchpad struct {
  ID          string    `gorm:"size:20;primaryKey"`
  Symbol      string    `gorm:"size:20;not null;index:idx_binance_spot_launchpad_symbol_status"`
  Capital     float64   `gorm:"not null"`
  Price       float64   `gorm:"not null"`
  CorePrice   float64   `gorm:"not null"`
  TakePrice   float64   `gorm:"not null"`
  StopPrice   float64   `gorm:"not null"`
  TakeOrderId int64     `gorm:"not null"`
  StopOrderId int64     `gorm:"not null"`
  Profit      float64   `gorm:"not null"`
  Timestamp   int64     `gorm:"not null"`
  Status      int       `gorm:"size:30;not null;index;index:idx_binance_spot_launchpad_symbol_status"`
  Version     int       `gorm:"not null"`
  Remark      string    `gorm:"size:5000;not null"`
  IssuedAt    time.Time `gorm:"not null"`
  ExpiredAt   time.Time `gorm:"not null"`
  CreatedAt   time.Time `gorm:"not null"`
  UpdatedAt   time.Time `gorm:"not null"`
}

func (m *Launchpad) TableName() string {
  return "binance_spot_launchpad"
}
