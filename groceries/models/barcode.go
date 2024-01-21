package models

import "time"

type Barcode struct {
  ID        string    `gorm:"size:20;primaryKey"`
  Uid       string    `gorm:"size:20;not null;uniqueIndex:unq_product_barcode"`
  Barcode   string    `gorm:"size:50;not null;uniqueIndex:unq_product_barcode"`
  ProductID string    `gorm:"size:20;not null"`
  CreatedAt time.Time `gorm:"not null"`
  UpdatedAt time.Time `gorm:"not null"`
}

func (m *Barcode) TableName() string {
  return "groceries_barcode"
}
