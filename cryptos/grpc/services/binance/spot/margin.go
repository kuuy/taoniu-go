package spot

import (
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/spot/margin"
)

type Margin struct {
  Db *gorm.DB
}

func NewMargin(db *gorm.DB) *Margin {
  return &Margin{
    Db: db,
  }
}

func (srv *Margin) Register(s *grpc.Server) error {
  margin.NewIsolated(srv.Db).Register(s)
  return nil
}
