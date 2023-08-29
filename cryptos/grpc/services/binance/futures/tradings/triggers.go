package tradings

import (
  "google.golang.org/grpc"
  "gorm.io/gorm"
)

type Triggers struct {
  Db *gorm.DB
}

func NewTriggers(db *gorm.DB) *Triggers {
  return &Triggers{
    Db: db,
  }
}

func (srv *Triggers) Register(s *grpc.Server) error {
  return nil
}
