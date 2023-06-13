package cross

import "time"

type Trigger struct {
  ID            string    `gorm:"size:20;primaryKey"`
  Symbol        string    `gorm:"size:20;not null;index:idx_binance_spot_margin_cross_triggers_symbol_status"`
  Capital       float64   `gorm:"not null"`
  Price         float64   `gorm:"not null"`
  EntryPrice    float64   `gorm:"not null"`
  EntryQuantity float64   `gorm:"not null"`
  Status        int       `gorm:"size:30;not null;index;index:idx_binance_spot_margin_cross_triggers_symbol_status"`
  Remark        string    `gorm:"size:5000;not null"`
  ExpiredAt     time.Time `gorm:"not null"`
  CreatedAt     time.Time `gorm:"not null"`
  UpdatedAt     time.Time `gorm:"not null"`
}

func (m *Trigger) TableName() string {
  return "binance_spot_margin_cross_triggers"
}
