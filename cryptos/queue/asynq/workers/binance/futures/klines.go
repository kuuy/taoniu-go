package futures

import (
  "context"
  "encoding/json"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type Klines struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.KlinesRepository
}

func NewKlines() *Klines {
  h := &Klines{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.KlinesRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  return h
}

type KlinesFlushPayload struct {
  Symbol   string
  Interval string
  Limit    int
  UseProxy bool
}

func (h *Klines) Flush(ctx context.Context, t *asynq.Task) error {
  var payload KlinesFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  if payload.UseProxy {
    h.Repository.UseProxy = true
  }

  h.Repository.Flush(payload.Symbol, payload.Interval, payload.Limit)

  return nil
}
