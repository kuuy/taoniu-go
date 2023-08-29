package spot

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type Indicators struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.IndicatorsRepository
}

type IndicatorPayload struct {
  Symbol   string
  Interval string
  Period   int
  Limit    int
}

type PivotPayload struct {
  Symbol   string
  Interval string
}

type KdjPayload struct {
  Symbol      string
  Interval    string
  LongPeriod  int
  ShortPeriod int
  Limit       int
}

type VolumeProfilePayload struct {
  Symbol   string
  Interval string
  Limit    int
}

func NewIndicators() *Indicators {
  h := &Indicators{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.IndicatorsRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
    Db: h.Db,
  }
  return h
}

func (h *Indicators) Pivot(ctx context.Context, t *asynq.Task) error {
  var payload PivotPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:pivot:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Pivot(payload.Symbol, payload.Interval)

  return nil
}

func (h *Indicators) Atr(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:atr:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Atr(payload.Symbol, payload.Interval, payload.Period, payload.Limit)

  return nil
}

func (h *Indicators) Zlema(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:zlema:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Zlema(payload.Symbol, payload.Interval, payload.Period, payload.Limit)

  return nil
}

func (h *Indicators) HaZlema(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:ha_zlema:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.HaZlema(payload.Symbol, payload.Interval, payload.Period, payload.Limit)

  return nil
}

func (h *Indicators) Kdj(ctx context.Context, t *asynq.Task) error {
  var payload KdjPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:kdj:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Kdj(payload.Symbol, payload.Interval, payload.LongPeriod, payload.ShortPeriod, payload.Limit)

  return nil
}

func (h *Indicators) BBands(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:bbands:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.BBands(payload.Symbol, payload.Interval, payload.Period, payload.Limit)

  return nil
}

func (h *Indicators) VolumeProfile(ctx context.Context, t *asynq.Task) error {
  var payload VolumeProfilePayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:volume_profile:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.VolumeProfile(payload.Symbol, payload.Interval, payload.Limit)

  return nil
}

func (h *Indicators) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:spot:indicators:atr", h.Atr)
  mux.HandleFunc("binance:spot:indicators:zlema", h.Zlema)
  mux.HandleFunc("binance:spot:indicators:ha_zlema", h.HaZlema)
  mux.HandleFunc("binance:spot:indicators:kdj", h.Kdj)
  mux.HandleFunc("binance:spot:indicators:bbands", h.BBands)
  mux.HandleFunc("binance:spot:indicators:pivot", h.Pivot)
  mux.HandleFunc("binance:spot:indicators:volume_profile", h.VolumeProfile)
  return nil
}
