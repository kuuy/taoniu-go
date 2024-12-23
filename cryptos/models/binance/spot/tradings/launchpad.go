package tradings

import "time"

type Launchpad struct {
  ID           string    `gorm:"size:20;primaryKey"`
  Symbol       string    `gorm:"size:20;not null"`
  LaunchpadID  string    `gorm:"size:20;index"`
  BuyPrice     float64   `gorm:"type:double precision;not null"`
  SellPrice    float64   `gorm:"type:double precision;not null"`
  BuyQuantity  float64   `gorm:"type:double precision;not null"`
  SellQuantity float64   `gorm:"type:double precision;not null"`
  BuyOrderId   int64     `gorm:"not null"`
  SellOrderId  int64     `gorm:"not null"`
  Status       int       `gorm:"not null;index:idx_binance_spot_tradings_launchpad_updated_status,priority:2"`
  Version      int       `gorm:"not null"`
  Remark       string    `gorm:"size:5000;not null"`
  CreatedAt    time.Time `gorm:"not null"`
  UpdatedAt    time.Time `gorm:"not null;index;index:idx_binance_spot_tradings_launchpad_updated_status,priority:1"`
}

func (m *Launchpad) TableName() string {
  return "binance_spot_tradings_launchpad"
}
