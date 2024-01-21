package tor

import "time"

type Bridge struct {
  ID           string    `gorm:"size:20;primaryKey"`
  Protocol     string    `gorm:"size:20;not null"`
  Ip           uint32    `gorm:"not null;uniqueIndex:unq_tor_bridges_ip_port"`
  Port         int       `gorm:"not null;uniqueIndex:unq_tor_bridges_ip_port"`
  Secret       string    `gorm:"size:40;not null"`
  Cert         string    `gorm:"size:100;not null"`
  Mode         int       `gorm:"not null"`
  TimeoutCount int       `gorm:"not null"`
  Status       int       `gorm:"not null;index"`
  CreatedAt    time.Time `gorm:"not null"`
  UpdatedAt    time.Time `gorm:"not null"`
}

func (m *Bridge) TableName() string {
  return "tor_bridges"
}
