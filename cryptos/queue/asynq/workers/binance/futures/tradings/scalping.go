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
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type Scalping struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.ScalpingRepository
}

func NewScalping() *Scalping {
  h := &Scalping{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.ScalpingRepository{
    Db: h.Db,
  }
  h.Repository.SymbolsRepository = &futuresRepositories.SymbolsRepository{
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

type ScalpingPlacePayload struct {
  PlanID string `json:"plan_id"`
}

type ScalpingFlushPayload struct {
  ID string `json:"id"`
}

func (h *Scalping) Place(ctx context.Context, t *asynq.Task) error {
  var payload ScalpingPlacePayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:futures:tradings:scalping:place:%s", payload.PlanID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Place(payload.PlanID)

  return nil
}

func (h *Scalping) Flush(ctx context.Context, t *asynq.Task) error {
  var payload ScalpingFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:futures:tradings:scalping:flush:%s", payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.ID)

  return nil
}