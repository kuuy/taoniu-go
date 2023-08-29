package tradings

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  futuresRepositories "taoniu.local/cryptos/repositories/dydx"
  repositories "taoniu.local/cryptos/repositories/dydx/tradings"
)

type Triggers struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.TriggersRepository
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
  h.Repository.MarketsRepository = &futuresRepositories.MarketsRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.AccountRepository = &futuresRepositories.AccountRepository{
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.OrdersRepository = &futuresRepositories.OrdersRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.PositionRepository = &futuresRepositories.PositionsRepository{
    Db: h.Db,
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

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:dydx:tradings:triggers:place:%s", payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Place(payload.ID)

  return nil
}

func (h *Triggers) Flush(ctx context.Context, t *asynq.Task) error {
  var payload TriggersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:dydx:tradings:triggers:flush:%s", payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.ID)

  return nil
}
