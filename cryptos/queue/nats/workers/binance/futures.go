package binance

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "gorm.io/gorm"
  "taoniu.local/cryptos/queue/nats/workers/binance/futures"
)

type Futures struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewFutures(db *gorm.DB, rdb *redis.Client, ctx context.Context) *Futures {
  return &Futures{
    Db:  db,
    Rdb: rdb,
    Ctx: ctx,
  }
}

func (h *Futures) Subscribe(nc *nats.Conn) error {
  futures.NewAccount(h.Rdb, h.Ctx).Subscribe(nc)
  futures.NewOrders(h.Db, h.Rdb, h.Ctx).Subscribe(nc)
  return nil
}
