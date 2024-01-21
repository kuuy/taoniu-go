package models

import "time"

type Apps struct {
  ID        string    `gorm:"size:20;primaryKey"`
  Phone     string    `gorm:"size:64;not null"`
  AppID     int       `gorm:"not null;uniqueIndex"`
  AppHash   string    `gorm:"size:64;not null"`
  Session   string    `gorm:"size:5000;not null"`
  Timestamp int64     `gorm:"not null"`
  Status    int       `gorm:"not null;index"`
  CreatedAt time.Time `gorm:"not null"`
  UpdatedAt time.Time `gorm:"not null"`
}

func (m *Apps) TableName() string {
  return "telegram_apps"
}
