package indicators

import (
  "context"
  "encoding/json"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type Daily struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.DailyRepository
}

func NewDaily() *Daily {
  h := &Daily{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.DailyRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  return h
}

func (h *Daily) Atr(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Atr(payload.Symbol, payload.Period, payload.Limit)

  return nil
}

func (h *Daily) Zlema(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Atr(payload.Symbol, payload.Period, payload.Limit)

  return nil
}

func (h *Daily) HaZlema(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.HaZlema(payload.Symbol, payload.Period, payload.Limit)

  return nil
}

func (h *Daily) Kdj(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Kdj(payload.Symbol, payload.LongPeriod, payload.ShortPeriod, payload.Limit)

  return nil
}

func (h *Daily) BBands(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.BBands(payload.Symbol, payload.Period, payload.Limit)

  return nil
}

func (h *Daily) Pivot(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.Pivot(payload.Symbol)

  return nil
}

func (h *Daily) VolumeProfile(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  h.Repository.VolumeProfile(payload.Symbol, payload.Limit)

  return nil
}

func (h *Daily) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:futures:indicators:1d:atr", h.Atr)
  mux.HandleFunc("binance:futures:indicators:1d:zlema", h.Zlema)
  mux.HandleFunc("binance:futures:indicators:1d:hazlema", h.HaZlema)
  mux.HandleFunc("binance:futures:indicators:1d:kdj", h.Kdj)
  mux.HandleFunc("binance:futures:indicators:1d:bbands", h.BBands)
  mux.HandleFunc("binance:futures:indicators:1d:pivot", h.Pivot)
  mux.HandleFunc("binance:futures:indicators:1d:volume_profile", h.VolumeProfile)
  return nil
}
