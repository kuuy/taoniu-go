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

type Strategies struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.StrategiesRepository
}

type StrategyPayload struct {
  Symbol   string
  Interval string
}

func NewStrategies() *Strategies {
  h := &Strategies{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.StrategiesRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.MarketsRepository = &repositories.MarketsRepository{
    Db: h.Db,
  }
  return h
}

func (h *Strategies) Atr(ctx context.Context, t *asynq.Task) error {
  var payload StrategyPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:dydx:strategies:atr:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Atr(payload.Symbol, payload.Interval)

  return nil
}

func (h *Strategies) Zlema(ctx context.Context, t *asynq.Task) error {
  var payload StrategyPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:dydx:strategies:zlema:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Atr(payload.Symbol, payload.Interval)

  return nil
}

func (h *Strategies) HaZlema(ctx context.Context, t *asynq.Task) error {
  var payload StrategyPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:dydx:strategies:ha_zlema:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.HaZlema(payload.Symbol, payload.Interval)

  return nil
}

func (h *Strategies) Kdj(ctx context.Context, t *asynq.Task) error {
  var payload StrategyPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:dydx:strategies:kdj:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Kdj(payload.Symbol, payload.Interval)

  return nil
}

func (h *Strategies) BBands(ctx context.Context, t *asynq.Task) error {
  var payload StrategyPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:dydx:strategies:bbands:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.BBands(payload.Symbol, payload.Interval)

  return nil
}

func (h *Strategies) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("dydx:strategies:atr", h.Atr)
  mux.HandleFunc("dydx:strategies:zlema", h.Zlema)
  mux.HandleFunc("dydx:strategies:ha_zlema", h.HaZlema)
  mux.HandleFunc("dydx:strategies:kdj", h.Kdj)
  mux.HandleFunc("dydx:strategies:bbands", h.BBands)
  return nil
}
