package proxies

import (
  "time"
)

type Proxy struct {
  ID           string    `gorm:"size:20;primaryKey"`
  Ip           uint32    `gorm:"not null;uniqueIndex:unq_proxies_ip_port"`
  Port         int       `gorm:"not null;uniqueIndex:unq_proxies_ip_port"`
  Timeout      float32   `gorm:"not null"`
  SuccessCount int       `gorm:"not null"`
  FailedCount  int       `gorm:"not null"`
  Timestamp    int64     `gorm:"not null;"`
  Status       int       `gorm:"not null;index"`
  Remark       string    `gorm:"size:5000;not null"`
  CreatedAt    time.Time `gorm:"not null"`
  UpdatedAt    time.Time `gorm:"not null"`
}

func (m *Proxy) TableName() string {
  return "proxies"
}
