package dydx

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
  "time"
)

type Orderbook struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.OrderbookRepository
}

func NewOrderbook(ansqContext *common.AnsqServerContext) *Orderbook {
  h := &Orderbook{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.OrderbookRepository{
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  return h
}

type OrderbookFlushPayload struct {
  Symbol   string
  UseProxy bool
}

func (h *Orderbook) Flush(ctx context.Context, t *asynq.Task) error {
  var payload OrderbookFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  if payload.UseProxy {
    h.Repository.UseProxy = true
  }

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:dydx:orderbook:%s:%t", payload.Symbol, payload.UseProxy),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.Symbol)

  return nil
}

func (h *Orderbook) Register() error {
  h.AnsqContext.Mux.HandleFunc("dydx:orderbook:flush", h.Flush)
  return nil
}
