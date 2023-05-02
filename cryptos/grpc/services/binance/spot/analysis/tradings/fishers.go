package tradings

import (
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/spot/analysis/tradings/fishers"
)

type Fishers struct {
  Db *gorm.DB
}

func NewFishers(db *gorm.DB) *Fishers {
  return &Fishers{
    Db: db,
  }
}

func (srv *Fishers) Register(s *grpc.Server) error {
  fishers.NewChart(srv.Db).Register(s)
  return nil
}
