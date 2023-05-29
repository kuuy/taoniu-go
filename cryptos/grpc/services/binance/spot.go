package binance

import (
  "context"
  "github.com/go-redis/redis/v8"
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/spot"
)

type Spot struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewSpot(
  db *gorm.DB,
  rdb *redis.Client,
  ctx context.Context,
) *Spot {
  return &Spot{
    Db:  db,
    Rdb: rdb,
    Ctx: ctx,
  }
}

func (srv *Spot) Register(s *grpc.Server) error {
  spot.NewMarkets(srv.Db, srv.Rdb, srv.Ctx).Register(s)
  spot.NewAnalysis(srv.Db).Register(s)
  spot.NewIndicators(srv.Db, srv.Rdb, srv.Ctx).Register(s)
  spot.NewTradings(srv.Db).Register(s)
  spot.NewMargin(srv.Db).Register(s)
  return nil
}
