package futures

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "strconv"
  "time"

  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
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
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  return h
}

func (h *Klines) Flush(ctx context.Context, t *asynq.Task) error {
  var payload KlinesFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_TASKS_KLINES_FLUSH, payload.Interval, payload.Symbol),
  )
  if !mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  timestamp := h.Repository.Timestamp(payload.Interval)
  redisKey := fmt.Sprintf(config.REDIS_KEY_KLINES, payload.Interval, payload.Symbol, timestamp)
  data, _ := h.AnsqContext.Rdb.HMGet(
    h.AnsqContext.Ctx,
    redisKey,
    "open",
    "close",
    "high",
    "low",
    "volume",
    "quota",
    "lasttime",
  ).Result()
  if data[0] != nil &&
    data[1] != nil &&
    data[2] != nil &&
    data[3] != nil &&
    data[4] != nil &&
    data[5] != nil &&
    data[6] != nil {
    open, _ := strconv.ParseFloat(data[0].(string), 64)
    close, _ := strconv.ParseFloat(data[1].(string), 64)
    high, _ := strconv.ParseFloat(data[2].(string), 64)
    low, _ := strconv.ParseFloat(data[3].(string), 64)
    volume, _ := strconv.ParseFloat(data[4].(string), 64)
    quota, _ := strconv.ParseFloat(data[5].(string), 64)
    lasttime, _ := strconv.ParseInt(data[6].(string), 10, 64)

    entity, err := h.Repository.Get(payload.Symbol, payload.Interval, timestamp)
    if errors.Is(err, gorm.ErrRecordNotFound) {
      h.Repository.Create(
        payload.Symbol,
        payload.Interval,
        open,
        close,
        high,
        low,
        volume,
        quota,
        timestamp,
      )
    } else if lasttime > entity.UpdatedAt.UnixMilli() {
      h.Repository.Updates(entity, map[string]interface{}{
        "open":   open,
        "close":  close,
        "high":   high,
        "low":    low,
        "volume": volume,
        "quota":  quota,
      })
    } else {
      diff := time.Now().UnixMilli() - entity.UpdatedAt.UnixMilli()
      if diff > 30000 {
        h.Repository.Flush(payload.Symbol, payload.Interval, 0, 1)
      }
    }
  } else {
    entity, err := h.Repository.Get(payload.Symbol, payload.Interval, timestamp)
    if errors.Is(err, gorm.ErrRecordNotFound) {
      h.Repository.Flush(payload.Symbol, payload.Interval, 0, 1)
    } else if time.Now().UnixMilli()-entity.UpdatedAt.UnixMilli() > 30000 {
      h.Repository.Flush(payload.Symbol, payload.Interval, 0, 1)
    }
  }

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
