package dydx

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type Plans struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.PlansRepository
}

type PlansPayload struct {
  Interval string
}

func NewPlans() *Plans {
  h := &Plans{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.PlansRepository{
    Db: h.Db,
  }
  h.Repository.MarketsRepository = &repositories.MarketsRepository{
    Db: h.Db,
  }
  return h
}

func (h *Plans) Flush(ctx context.Context, t *asynq.Task) error {
  var payload PlansPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:dydx:plans:%s:flush", payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.Interval)

  return nil
}

func (h *Plans) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("dydx:plans:flush", h.Flush)
  return nil
}
