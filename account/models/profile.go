package models

import "time"

type Profile struct {
  UID       string    `gorm:"size:20;primaryKey"`
  Nickname  string    `gorm:"size:64;not null;uniqueIndex"`
  Avatar    string    `gorm:"size:128;not null"`
  CreatedAt time.Time `gorm:"not null"`
  UpdatedAt time.Time `gorm:"not null"`
}

func (m *Profile) TableName() string {
  return "profile"
}
