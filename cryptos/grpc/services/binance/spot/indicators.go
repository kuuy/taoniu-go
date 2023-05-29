package spot

import (
  "context"
  "github.com/go-redis/redis/v8"
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/spot/indicators"
)

type Indicators struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewIndicators(db *gorm.DB, rdb *redis.Client, ctx context.Context) *Indicators {
  return &Indicators{
    Db:  db,
    Rdb: rdb,
    Ctx: ctx,
  }
}

func (srv *Indicators) Register(s *grpc.Server) error {
  indicators.NewDaily(srv.Db, srv.Rdb, srv.Ctx).Register(s)
  return nil
}
