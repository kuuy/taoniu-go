package dydx

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type Strategies struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.StrategiesRepository
}

type StrategyPayload struct {
  Symbol   string
  Interval string
}

func NewStrategies(ansqContext *common.AnsqServerContext) *Strategies {
  h := &Strategies{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.StrategiesRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.MarketsRepository = &repositories.MarketsRepository{
    Db: h.AnsqContext.Db,
  }
  return h
}

func (h *Strategies) Atr(ctx context.Context, t *asynq.Task) error {
  var payload StrategyPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
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
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
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
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
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
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
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
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:dydx:strategies:bbands:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.BBands(payload.Symbol, payload.Interval)

  return nil
}

func (h *Strategies) Register() error {
  h.AnsqContext.Mux.HandleFunc("dydx:strategies:atr", h.Atr)
  h.AnsqContext.Mux.HandleFunc("dydx:strategies:zlema", h.Zlema)
  h.AnsqContext.Mux.HandleFunc("dydx:strategies:ha_zlema", h.HaZlema)
  h.AnsqContext.Mux.HandleFunc("dydx:strategies:kdj", h.Kdj)
  h.AnsqContext.Mux.HandleFunc("dydx:strategies:bbands", h.BBands)
  return nil
}
