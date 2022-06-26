package models

import (
	"time"
)

type Future struct {
  ID string `gorm:"size:20;primaryKey"`
  Symbol string `gorm:"size:20;not null;unique"`
  Price float64 `gorm:"type:float;not null"`
  Open float64 `gorm:"type:float;not null"`
  High float64 `gorm:"type:float;not null"`
  Low float64 `gorm:"type:float;not null"`
  Volume float64 `gorm:"type:float;not null"`
  Quota float64 `gorm:"type:float;not null"`
  TicketStep float64 `gorm:"type:float;not null"`
  CreatedAt time.Time `gorm:"type:time;not null"`
  UpdatedAt time.Time `gorm:"type:time;not null"`
}

