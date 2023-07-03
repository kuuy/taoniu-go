package services

import (
  "context"
  "github.com/go-redis/redis/v8"
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance"
)

type Binance struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewBinance(
  db *gorm.DB,
  rdb *redis.Client,
  ctx context.Context,
) *Binance {
  return &Binance{
    Db:  db,
    Rdb: rdb,
    Ctx: ctx,
  }
}

func (srv *Binance) Register(s *grpc.Server) error {
  binance.NewSpot(srv.Db, srv.Rdb, srv.Ctx).Register(s)
  binance.NewFutures(srv.Db, srv.Rdb, srv.Ctx).Register(s)
  return nil
}
