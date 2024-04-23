package spot

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "time"

  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type Klines struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.KlinesRepository
}

func NewKlines(ansqContext *common.AnsqServerContext) *Klines {
  h := &Klines{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.KlinesRepository{
    Db:   h.AnsqContext.Db,
    Rdb:  h.AnsqContext.Rdb,
    Ctx:  h.AnsqContext.Ctx,
    Nats: h.AnsqContext.Nats,
  }
  return h
}

func (h *Klines) Flush(ctx context.Context, t *asynq.Task) error {
  var payload KlinesFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  if payload.UseProxy {
    h.Repository.UseProxy = true
  }

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_KLINES_FLUSH, payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.Symbol, payload.Interval, 0, payload.Limit)

  return nil
}

func (h *Klines) Update(ctx context.Context, t *asynq.Task) error {
  var payload KlinesUpdatePayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_KLINES_UPDATE, payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(10 * time.Second) {
    return nil
  }

  kline, err := h.Repository.Get(payload.Symbol, payload.Interval, payload.Timestamp)
  if errors.Is(err, gorm.ErrRecordNotFound) {
    h.Repository.Create(
      payload.Symbol,
      payload.Interval,
      payload.Open,
      payload.Close,
      payload.High,
      payload.Low,
      payload.Volume,
      payload.Quota,
      payload.Timestamp,
    )
  } else {
    h.Repository.Updates(kline, map[string]interface{}{
      "open":   payload.Open,
      "close":  payload.Close,
      "high":   payload.High,
      "low":    payload.Low,
      "volume": payload.Volume,
      "quota":  payload.Quota,
    })
  }

  message, _ := json.Marshal(map[string]interface{}{
    "symbol":   payload.Symbol,
    "interval": payload.Interval,
  })
  h.AnsqContext.Nats.Publish(config.NATS_KLINES_UPDATE, message)
  h.AnsqContext.Nats.Flush()

  return nil
}

func (h *Klines) Clean(ctx context.Context, t *asynq.Task) error {
  var payload KlinesCleanPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_KLINES_CLEAN, payload.Symbol),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Clean(payload.Symbol)

  return nil
}

func (h *Klines) Register() error {
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_KLINES_FLUSH, h.Flush)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_KLINES_UPDATE, h.Update)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_KLINES_CLEAN, h.Clean)
  return nil
}
