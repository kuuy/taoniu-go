package workers

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "gorm.io/gorm"
  "taoniu.local/cryptos/queue/nats/workers/binance"
)

type Binance struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewBinance(db *gorm.DB, rdb *redis.Client, ctx context.Context) *Binance {
  return &Binance{
    Db:  db,
    Rdb: rdb,
    Ctx: ctx,
  }
}

func (h *Binance) Subscribe(nc *nats.Conn) error {
  binance.NewSpot(h.Db, h.Rdb, h.Ctx).Subscribe(nc)
  binance.NewFutures(h.Db, h.Rdb, h.Ctx).Subscribe(nc)
  return nil
}
