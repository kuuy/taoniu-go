package proxies

import (
	"gorm.io/datatypes"
	"time"
)

type Source struct {
	ID        string            `gorm:"size:20;primaryKey"`
	Url       string            `gorm:"size:155;not null;"`
	UrlSha1   string            `gorm:"size:40;not null;index"`
	Headers   datatypes.JSONMap `gorm:"not null"`
	UseProxy  bool              `gorm:"not null"`
	Timeout   int               `gorm:"not null"`
	Rules     datatypes.JSONMap `gorm:"not null"`
	Status    int               `gorm:"not null;index"`
	Remark    string            `gorm:"size:5000;not null"`
	CreatedAt time.Time         `gorm:"not null"`
	UpdatedAt time.Time         `gorm:"not null"`
}

func (m *Source) TableName() string {
	return "proxies_sources"
}
