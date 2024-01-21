package models

import (
  "time"
)

type Store struct {
  ID        string    `gorm:"size:20;primaryKey"`
  Uid       string    `gorm:"size:20;not null"`
  Name      string    `gorm:"size:50;not null"`
  Logo      string    `gorm:"size:155;not null"`
  CreatedAt time.Time `gorm:"not null"`
  UpdatedAt time.Time `gorm:"not null"`
}

func (m *Store) TableName() string {
  return "groceries_stores"
}
