package dydx

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "time"

  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
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
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
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

func (h *Klines) Register() error {
  h.AnsqContext.Mux.HandleFunc("dydx:klines:flush", h.Flush)
  return nil
}
