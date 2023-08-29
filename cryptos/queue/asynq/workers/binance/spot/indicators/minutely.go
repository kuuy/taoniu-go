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

func (h *Minutely) Pivot(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1m:pivot"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Pivot(payload.Symbol)

  return nil
}

func (h *Minutely) Atr(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1m:atr"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Atr(payload.Symbol, payload.Period, payload.Limit)

  return nil
}

func (h *Minutely) Zlema(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1m:zlema"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Zlema(payload.Symbol, payload.Period, payload.Limit)

  return nil
}

func (h *Minutely) HaZlema(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1m:hazlema"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.HaZlema(payload.Symbol, payload.Period, payload.Limit)

  return nil
}

func (h *Minutely) Kdj(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1m:kdj"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Kdj(payload.Symbol, payload.LongPeriod, payload.ShortPeriod, payload.Limit)

  return nil
}

func (h *Minutely) BBands(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1m:bbands"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.BBands(payload.Symbol, payload.Period, payload.Limit)

  return nil
}

func (h *Minutely) VolumeProfile(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:1m:volume_profile"),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.VolumeProfile(payload.Symbol, payload.Limit)

  return nil
}

func (h *Minutely) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:spot:indicators:1m:pivot", h.Pivot)
  mux.HandleFunc("binance:spot:indicators:1m:atr", h.Atr)
  mux.HandleFunc("binance:spot:indicators:1m:zlema", h.Zlema)
  mux.HandleFunc("binance:spot:indicators:1m:hazlema", h.HaZlema)
  mux.HandleFunc("binance:spot:indicators:1m:kdj", h.Kdj)
  mux.HandleFunc("binance:spot:indicators:1m:bbands", h.BBands)
  mux.HandleFunc("binance:spot:indicators:1m:volume_profile", h.VolumeProfile)
  return nil
}
