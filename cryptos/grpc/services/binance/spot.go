package binance

import (
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/spot"
)

type Spot struct {
  Db *gorm.DB
}

func NewSpot(db *gorm.DB) *Spot {
  return &Spot{
    Db: db,
  }
}

func (srv *Spot) Register(s *grpc.Server) error {
  spot.NewAnalysis(srv.Db).Register(s)
  return nil
}
