package isolated

import (
  "google.golang.org/grpc"
  "gorm.io/gorm"
)

type Tradings struct {
  Db *gorm.DB
}

func NewTradings(db *gorm.DB) *Tradings {
  return &Tradings{
    Db: db,
  }
}

func (srv *Tradings) Register(s *grpc.Server) error {
  return nil
}
