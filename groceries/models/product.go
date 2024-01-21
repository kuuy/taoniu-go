package models

import (
  "time"
)

type Product struct {
  ID        string    `gorm:"size:20;primaryKey"`
  Uid       string    `gorm:"size:20;not null;index"`
  Title     string    `gorm:"size:50;not null"`
  Intro     string    `gorm:"size:5000;not null"`
  Price     float64   `gorm:"not null"`
  Cover     string    `gorm:"size:155;not null"`
  CreatedAt time.Time `gorm:"not null"`
  UpdatedAt time.Time `gorm:"not null"`
}

func (m *Product) TableName() string {
  return "groceries_products"
}
