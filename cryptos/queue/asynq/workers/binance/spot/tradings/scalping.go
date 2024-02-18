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
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type Scalping struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.ScalpingRepository
}

func NewScalping() *Scalping {
  h := &Scalping{
    Db:  common.NewDB(1),
    Rdb: common.NewRedis(1),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.ScalpingRepository{
    Db: h.Db,
  }
  h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.AccountRepository = &spotRepositories.AccountRepository{
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.PositionRepository = &spotRepositories.PositionsRepository{
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
    fmt.Sprintf("locks:binance:spot:tradings:scalping:place:%s", payload.PlanID),
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
    fmt.Sprintf("locks:binance:spot:tradings:scalping:flush:%s", payload.ID),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.ID)

  return nil
}
