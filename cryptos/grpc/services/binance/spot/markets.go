package spot

import (
  "context"
  "github.com/go-redis/redis/v8"
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/spot/markets"
)

type Markets struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewMarkets(
  db *gorm.DB,
  rdb *redis.Client,
  ctx context.Context,
) *Markets {
  return &Markets{
    Db:  db,
    Rdb: rdb,
    Ctx: ctx,
  }
}

func (srv *Markets) Register(s *grpc.Server) error {
  markets.NewLive(srv.Db, srv.Rdb, srv.Ctx).Register(s)
  return nil
}
