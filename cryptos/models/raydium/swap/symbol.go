package swap

import (
  "time"
)

type Symbol struct {
  ID           string    `gorm:"size:20;primaryKey"`
  Symbol       string    `gorm:"size:20;not null;uniqueIndex"`
  BaseAddress  string    `gorm:"not null"`
  QuoteAddress string    `gorm:"not null"`
  Status       int       `gorm:"type:integer;not null;index"`
  CreatedAt    time.Time `gorm:"not null"`
  UpdatedAt    time.Time `gorm:"not null"`
}

func (m *Symbol) TableName() string {
  return "raydium_swap_symbols"
}
