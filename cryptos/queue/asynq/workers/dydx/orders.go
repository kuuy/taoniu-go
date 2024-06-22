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

type Orders struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.OrdersRepository
}

func NewOrders(ansqContext *common.AnsqServerContext) *Orders {
  h := &Orders{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.OrdersRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  return h
}

type OrdersOpenPayload struct {
  Symbol string `json:"symbol"`
}

type OrdersFlushPayload struct {
  OrderId string `json:"order_id"`
}

func (h *Orders) Open(ctx context.Context, t *asynq.Task) error {
  var payload OrdersOpenPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:dydx:orders:open:%s", payload.Symbol),
  )
  if !mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Open(payload.Symbol)

  return nil
}

func (h *Orders) Flush(ctx context.Context, t *asynq.Task) error {
  var payload OrdersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:dydx:orders:flush:%d", payload.OrderId),
  )
  if !mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.OrderId)

  return nil
}

func (h *Orders) Register() error {
  h.AnsqContext.Mux.HandleFunc("dydx:orders:open", h.Open)
  h.AnsqContext.Mux.HandleFunc("dydx:orders:flush", h.Flush)
  return nil
}
