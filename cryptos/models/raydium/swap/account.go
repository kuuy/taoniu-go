package swap

import (
  "time"
)

type Account struct {
  ID        string    `gorm:"size:20;primaryKey"`
  Symbol    string    `gorm:"size:20;not null;index"`
  Address   string    `gorm:"size:44;not null;uniqueIndex"`
  Balance   float64   `gorm:"type:double precision;not null"`
  Status    int       `gorm:"type:integer;not null;index"`
  CreatedAt time.Time `gorm:"not null"`
  UpdatedAt time.Time `gorm:"not null"`
}

func (m *Account) TableName() string {
  return "raydium_swap_accounts"
}
