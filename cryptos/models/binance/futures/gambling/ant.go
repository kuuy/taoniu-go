package gambling

import (
  "time"
)

type Ant struct {
  ID            string    `gorm:"size:20;primaryKey"`
  Symbol        string    `gorm:"size:20;not null;index:idx_binance_futures_gambling_ant"`
  Side          int       `gorm:"not null"`
  EntryPrice    float64   `gorm:"not null"`
  EntryQuantity float64   `gorm:"not null"`
  PlaceQuantity float64   `gorm:"not null"`
  TakeQuantity  float64   `gorm:"not null"`
  Profit        float64   `gorm:"not null"`
  Timestamp     int64     `gorm:"not null"`
  Status        int       `gorm:"not null;index;index:idx_binance_futures_gambling_ant"`
  Version       int       `gorm:"not null"`
  Remark        string    `gorm:"size:5000;not null"`
  ExpiredAt     time.Time `gorm:"not null"`
  CreatedAt     time.Time `gorm:"not null"`
  UpdatedAt     time.Time `gorm:"not null"`
}

func (m *Ant) TableName() string {
  return "binance_futures_gambling_ant"
}
