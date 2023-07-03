package tradings

import (
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/futures/tradings/triggers"
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
  triggers.NewGrids(srv.Db).Register(s)
  return nil
}
