package spot

import (
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/spot/tradings"
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
  tradings.NewFishers(srv.Db).Register(s)
  tradings.NewScalping(srv.Db).Register(s)
  tradings.NewTriggers(srv.Db).Register(s)
  return nil
}
