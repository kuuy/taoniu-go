package spot

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/common"

  repositories "taoniu.local/cryptos/repositories/binance/spot"
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
  Symbol  string `json:"symbol"`
  OrderID int64  `json:"order_id"`
}

type OrdersSyncPayload struct {
  Symbol    string `json:"symbol"`
  StartTime int64  `json:"start_time"`
  limit     int    `json:"limit"`
}

func (h *Orders) Open(ctx context.Context, t *asynq.Task) error {
  var payload OrdersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:orders:open:%s", payload.Symbol),
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
    fmt.Sprintf("locks:binance:spot:orders:flush:%s:%d", payload.Symbol, payload.OrderID),
  )
  if !mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.Symbol, payload.OrderID)

  return nil
}

func (h *Orders) Sync(ctx context.Context, t *asynq.Task) error {
  var payload OrdersSyncPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:orders:sync:%s", payload.Symbol),
  )
  if !mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Sync(payload.Symbol, payload.StartTime, payload.limit)

  return nil
}

func (h *Orders) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:spot:orders:open", h.Open)
  h.AnsqContext.Mux.HandleFunc("binance:spot:orders:flush", h.Flush)
  h.AnsqContext.Mux.HandleFunc("binance:spot:orders:sync", h.Sync)
  return nil
}
