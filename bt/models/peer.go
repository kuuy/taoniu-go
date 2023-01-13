package models

import (
	"gorm.io/datatypes"
	"time"
)

type Peer struct {
	ID                 string         `gorm:"size:20;primaryKey"`
	Tid                string         `gorm:"size:4;not null;index:idx_bt_peers_tid_updated,priority:2"`
	Infohash           string         `gorm:"size:40;not null;uniqueIndex"`
	Sources            datatypes.JSON `gorm:"not null"`
	Data               datatypes.JSON `gorm:"not null"`
	TimeoutCount       int            `gorm:"not null"`
	StreakTimeoutCount int            `gorm:"not null"`
	Status             int            `gorm:"not null;index"`
	CreatedAt          time.Time      `gorm:"not null"`
	UpdatedAt          time.Time      `gorm:"not null;index:idx_bt_peers_tid_updated,priority:1"`
}

func (m *Peer) TableName() string {
	return "bt_peers"
}
