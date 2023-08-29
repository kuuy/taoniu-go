package indicators

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"
  "taoniu.local/cryptos/common"

  repositories "taoniu.local/cryptos/repositories/binance/spot/indicators"
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

func (h *Daily) Pivot(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1d:pivot"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Pivot(payload.Symbol)

  return nil
}

func (h *Daily) Atr(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1d:atr"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Atr(payload.Symbol, payload.Period, payload.Limit)

  return nil
}

func (h *Daily) Zlema(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1d:zlema"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Zlema(payload.Symbol, payload.Period, payload.Limit)

  return nil
}

func (h *Daily) HaZlema(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1d:hazlema"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.HaZlema(payload.Symbol, payload.Period, payload.Limit)

  return nil
}

func (h *Daily) Kdj(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1d:kdj"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Kdj(payload.Symbol, payload.LongPeriod, payload.ShortPeriod, payload.Limit)

  return nil
}

func (h *Daily) BBands(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1d:bbands"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.BBands(payload.Symbol, payload.Period, payload.Limit)

  return nil
}

func (h *Daily) VolumeProfile(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1d:volume_profile"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.VolumeProfile(payload.Symbol, payload.Limit)

  return nil
}

func (h *Daily) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:spot:indicators:1d:atr", h.Atr)
  mux.HandleFunc("binance:spot:indicators:1d:zlema", h.Zlema)
  mux.HandleFunc("binance:spot:indicators:1d:hazlema", h.HaZlema)
  mux.HandleFunc("binance:spot:indicators:1d:kdj", h.Kdj)
  mux.HandleFunc("binance:spot:indicators:1d:bbands", h.BBands)
  mux.HandleFunc("binance:spot:indicators:1d:pivot", h.Pivot)
  mux.HandleFunc("binance:spot:indicators:1d:volume_profile", h.VolumeProfile)
  return nil
}
