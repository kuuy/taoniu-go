package margin

import (
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/margin/isolated"
)

type Isolated struct {
  Db *gorm.DB
}

func NewIsolated(db *gorm.DB) *Isolated {
  return &Isolated{
    Db: db,
  }
}

func (srv *Isolated) Register(s *grpc.Server) error {
  isolated.NewTradings(srv.Db).Register(s)
  return nil
}
