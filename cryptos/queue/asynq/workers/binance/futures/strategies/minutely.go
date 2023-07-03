package strategies

import (
  "context"
  "encoding/json"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures/strategies"
)

type Minutely struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.MinutelyRepository
}

func NewMinutely() *Minutely {
  h := &Minutely{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.MinutelyRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  return h
}

type MinutelyFlushPayload struct {
  Symbol string
  Period int
  Limit  int
}

func (h *Minutely) Atr(ctx context.Context, t *asynq.Task) error {
  var payload MinutelyFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Atr(payload.Symbol)

  return nil
}

func (h *Minutely) Zlema(ctx context.Context, t *asynq.Task) error {
  var payload MinutelyFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Atr(payload.Symbol)

  return nil
}

func (h *Minutely) HaZlema(ctx context.Context, t *asynq.Task) error {
  var payload MinutelyFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.HaZlema(payload.Symbol)

  return nil
}

func (h *Minutely) Kdj(ctx context.Context, t *asynq.Task) error {
  var payload MinutelyFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Kdj(payload.Symbol)

  return nil
}

func (h *Minutely) BBands(ctx context.Context, t *asynq.Task) error {
  var payload MinutelyFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.BBands(payload.Symbol)

  return nil
}

func (h *Minutely) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:futures:strategies:1m:atr", h.Atr)
  mux.HandleFunc("binance:futures:strategies:1m:zlema", h.Zlema)
  mux.HandleFunc("binance:futures:strategies:1m:hazlema", h.HaZlema)
  mux.HandleFunc("binance:futures:strategies:1m:kdj", h.Kdj)
  mux.HandleFunc("binance:futures:strategies:1m:bbands", h.BBands)
  return nil
}
