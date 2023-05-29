package indicators

import (
  "context"
  "github.com/go-redis/redis/v8"
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/spot/indicators/daily"
)

type Daily struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewDaily(db *gorm.DB, rdb *redis.Client, ctx context.Context) *Daily {
  return &Daily{
    Db:  db,
    Rdb: rdb,
    Ctx: ctx,
  }
}

func (srv *Daily) Register(s *grpc.Server) error {
  daily.NewRanking(srv.Db, srv.Rdb, srv.Ctx).Register(s)
  return nil
}
