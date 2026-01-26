package swap

import (
  "gorm.io/datatypes"
  "time"
)

type Mint struct {
  ID        string                      `gorm:"size:20;primaryKey"`
  Name      string                      `gorm:"size:50;not null"`
  Symbol    string                      `gorm:"size:20;not null;uniqueIndex"`
  Address   string                      `gorm:"size:44;not null"`
  Decimals  int                         `gorm:"type:integer;not null"`
  Tags      datatypes.JSONSlice[string] `gorm:"size:155;not null"`
  Status    int                         `gorm:"type:integer;not null;index"`
  CreatedAt time.Time                   `gorm:"not null"`
  UpdatedAt time.Time                   `gorm:"not null"`
}

func (m *Mint) TableName() string {
  return "raydium_swap_mints"
}
