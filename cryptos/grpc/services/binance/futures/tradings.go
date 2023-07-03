package futures

import (
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/futures/tradings"
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
  tradings.NewTriggers(srv.Db).Register(s)
  return nil
}
