package services

import (
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance"
)

type Binance struct {
  Db *gorm.DB
}

func NewBinance(db *gorm.DB) *Binance {
  return &Binance{
    Db: db,
  }
}

func (srv *Binance) Register(s *grpc.Server) error {
  binance.NewSpot(srv.Db).Register(s)
  return nil
}
