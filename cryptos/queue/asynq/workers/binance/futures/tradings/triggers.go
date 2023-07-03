package tradings

import (
  "context"
  "encoding/json"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type Triggers struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Repository        *repositories.TriggersRepository
  SymbolsRepository *futuresRepositories.SymbolsRepository
}

func NewTriggers() *Triggers {
  h := &Triggers{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.TriggersRepository{
    Db: h.Db,
  }
  h.Repository.PositionRepository = &futuresRepositories.PositionsRepository{
    Db: h.Db,
  }
  h.Repository.SymbolsRepository = &futuresRepositories.SymbolsRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.OrdersRepository = &futuresRepositories.OrdersRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  return h
}

type TriggersPlacePayload struct {
  ID string
}

type TriggersFlushPayload struct {
  ID string
}

func (h *Triggers) Place(ctx context.Context, t *asynq.Task) error {
  var payload TriggersPlacePayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Place(payload.ID)

  return nil
}

func (h *Triggers) Flush(ctx context.Context, t *asynq.Task) error {
  var payload TriggersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Flush(payload.ID)

  return nil
}
