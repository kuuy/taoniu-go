package spot

import (
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/spot/analysis"
)

type Analysis struct {
  Db *gorm.DB
}

func NewAnalysis(db *gorm.DB) *Analysis {
  return &Analysis{
    Db: db,
  }
}

func (srv *Analysis) Register(s *grpc.Server) error {
  analysis.NewTradings(srv.Db).Register(s)
  return nil
}
