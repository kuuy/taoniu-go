package futures

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
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
  OrderId int64  `json:"order_id"`
}

type OrdersSyncPayload struct {
  Symbol    string `json:"symbol"`
  StartTime int64  `json:"start_time"`
  limit     int    `json:"limit"`
}

func (h *Orders) Open(ctx context.Context, t *asynq.Task) error {
  var payload OrdersOpenPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_ORDERS_OPEN, payload.Symbol),
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
    fmt.Sprintf(config.LOCKS_ORDERS_FLUSH, payload.Symbol, payload.OrderId),
  )
  if !mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.Symbol, payload.OrderId)

  return nil
}

func (h *Orders) Sync(ctx context.Context, t *asynq.Task) error {
  var payload OrdersSyncPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf(config.LOCKS_ORDERS_SYNC, payload.Symbol),
  )
  if !mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Sync(payload.Symbol, payload.StartTime, payload.limit)

  return nil
}

func (h *Orders) Register() error {
  h.AnsqContext.Mux.HandleFunc("binance:futures:orders:open", h.Open)
  h.AnsqContext.Mux.HandleFunc("binance:futures:orders:flush", h.Flush)
  h.AnsqContext.Mux.HandleFunc("binance:futures:orders:sync", h.Sync)
  return nil
}
