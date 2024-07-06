package margin

import (
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/margin/cross"
)

type Cross struct {
  Db *gorm.DB
}

func NewCross(db *gorm.DB) *Cross {
  return &Cross{
    Db: db,
  }
}

func (srv *Cross) Register(s *grpc.Server) error {
  cross.NewTradings(srv.Db).Register(s)
  return nil
}
