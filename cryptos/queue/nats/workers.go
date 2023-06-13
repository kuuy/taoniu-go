package nats

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "gorm.io/gorm"
  "taoniu.local/cryptos/queue/nats/workers"
)

type Workers struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewWorkers(db *gorm.DB, rdb *redis.Client, ctx context.Context) *Workers {
  return &Workers{
    Db:  db,
    Rdb: rdb,
    Ctx: ctx,
  }
}

func (h *Workers) Subscribe(nc *nats.Conn) error {
  workers.NewBinance(h.Db, h.Rdb, h.Ctx).Subscribe(nc)
  return nil
}
