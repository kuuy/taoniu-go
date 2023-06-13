package spot

import "time"

type Triggers struct {
  ID            string    `gorm:"size:20;primaryKey"`
  Symbol        string    `gorm:"size:20;not null;index:idx_binance_spot_triggers_symbol_status"`
  Capital       float64   `gorm:"not null"`
  Multiple      int       `gorm:"not null"`
  Price         float64   `gorm:"not null"`
  EntryPrice    float64   `gorm:"not null"`
  EntryQuantity float64   `gorm:"not null"`
  Status        int       `gorm:"size:30;not null;index;index:idx_binance_spot_triggers_symbol_status"`
  Remark        string    `gorm:"size:5000;not null"`
  ExpiredAt     time.Time `gorm:"not null"`
  CreatedAt     time.Time `gorm:"not null"`
  UpdatedAt     time.Time `gorm:"not null"`
}

func (m *Triggers) TableName() string {
  return "binance_spot_triggers"
}