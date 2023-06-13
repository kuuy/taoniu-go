package binance

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "gorm.io/gorm"
  "taoniu.local/cryptos/queue/nats/workers/binance/spot"
)

type Spot struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewSpot(db *gorm.DB, rdb *redis.Client, ctx context.Context) *Spot {
  return &Spot{
    Db:  db,
    Rdb: rdb,
    Ctx: ctx,
  }
}

func (h *Spot) Subscribe(nc *nats.Conn) error {
  spot.NewAccount(h.Rdb, h.Ctx).Subscribe(nc)
  spot.NewTickers(h.Rdb, h.Ctx).Subscribe(nc)
  return nil
}
