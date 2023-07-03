package binance

import (
  "context"
  "github.com/go-redis/redis/v8"
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/futures"
)

type Futures struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewFutures(
  db *gorm.DB,
  rdb *redis.Client,
  ctx context.Context,
) *Futures {
  return &Futures{
    Db:  db,
    Rdb: rdb,
    Ctx: ctx,
  }
}

func (srv *Futures) Register(s *grpc.Server) error {
  futures.NewMarkets(srv.Db, srv.Rdb, srv.Ctx).Register(s)
  futures.NewIndicators(srv.Db, srv.Rdb, srv.Ctx).Register(s)
  futures.NewPlans(srv.Db).Register(s)
  futures.NewTradings(srv.Db).Register(s)
  return nil
}
