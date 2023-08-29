package dydx

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/nats-io/nats.go"
  "log"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type Klines struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Nats       *nats.Conn
  Repository *repositories.KlinesRepository
}

func NewKlines() *Klines {
  h := &Klines{
    Db:   common.NewDB(),
    Rdb:  common.NewRedis(),
    Ctx:  context.Background(),
    Nats: common.NewNats(),
  }
  h.Repository = &repositories.KlinesRepository{
    Db:   h.Db,
    Rdb:  h.Rdb,
    Ctx:  h.Ctx,
    Nats: h.Nats,
  }
  return h
}

type KlinesFlushPayload struct {
  Symbol   string
  Interval string
  Endtime  int64
  Limit    int
  UseProxy bool
}

func (h *Klines) Flush(ctx context.Context, t *asynq.Task) error {
  var payload KlinesFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  if payload.UseProxy {
    h.Repository.UseProxy = true
  }

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:dydx:klines:%s:%s:%t", payload.Symbol, payload.Interval, payload.UseProxy),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  err := h.Repository.Flush(payload.Symbol, payload.Interval, payload.Endtime, payload.Limit)
  if err != nil {
    log.Println("klines flush error", err.Error())
  }

  return nil
}

func (h *Klines) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("dydx:klines:flush", h.Flush)
  return nil
}
