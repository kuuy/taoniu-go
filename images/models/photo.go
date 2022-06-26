package models

import (
	"time"
)

type Photo struct {
  ID string `gorm:"size:20;primaryKey"`
  Title string `gorm:"size:50;not null"`
  Intro string `gorm:"size:5000;not null"`
  Width int64 `gorm:"not null;"`
  Height int64 `gorm:"not null;"`
  Mime string `gorm:"not null;"`
  Size int64 `gorm:"not null;"`
  Filepath string `gorm:"size:64;not null;unique"`
  Filename string `gorm:"size:64;not null;unique"`
  Filehash string `gorm:"size:64;not null;unique"`
  CreatedAt time.Time `gorm:"not null"`
  UpdatedAt time.Time `gorm:"not null"`
}

