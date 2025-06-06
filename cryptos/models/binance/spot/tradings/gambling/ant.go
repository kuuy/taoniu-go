package gambling

import "time"

type Ant struct {
  ID        string    `gorm:"size:20;primaryKey"`
  Symbol    string    `gorm:"size:20;not null"`
  AntId     string    `gorm:"size:20;index:idx_binance_spot_tradings_gambling_ant"`
  Mode      int       `gorm:"type:integer;not null;index:idx_binance_spot_tradings_gambling_ant"`
  Price     float64   `gorm:"type:double precision;not null"`
  Quantity  float64   `gorm:"type:double precision;not null"`
  OrderId   int64     `gorm:"not null"`
  Status    int       `gorm:"type:integer;not null;index;index:idx_binance_spot_tradings_gambling_ant"`
  Version   int       `gorm:"not null"`
  Remark    string    `gorm:"size:5000;not null"`
  CreatedAt time.Time `gorm:"not null"`
  UpdatedAt time.Time `gorm:"not null"`
}

func (m *Ant) TableName() string {
  return "binance_spot_tradings_gambling_ant"
}
