package indicators

import (
  "context"
  "github.com/go-redis/redis/v8"
  "google.golang.org/grpc"
  "gorm.io/gorm"
  "taoniu.local/cryptos/grpc/services/binance/futures/indicators/daily"
)

type Minutely struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewMinutely(db *gorm.DB, rdb *redis.Client, ctx context.Context) *Minutely {
  return &Minutely{
    Db:  db,
    Rdb: rdb,
    Ctx: ctx,
  }
}

func (srv *Minutely) Register(s *grpc.Server) error {
  daily.NewRanking(srv.Db, srv.Rdb, srv.Ctx).Register(s)
  return nil
}
