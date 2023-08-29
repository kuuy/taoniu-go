package spot

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
  "time"
)

type Orders struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.OrdersRepository
}

func NewOrders() *Orders {
  h := &Orders{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.OrdersRepository{
    Db:  h.Db,
    Rdb: h.Rdb,
    Ctx: h.Ctx,
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
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:tradings:orders:open:%s", payload.Symbol),
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
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:tradings:orders:flush:%s:%d", payload.Symbol, payload.OrderID),
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
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:spot:tradings:orders:sync:%s", payload.Symbol),
  )
  if !mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Sync(payload.Symbol, payload.StartTime, payload.limit)

  return nil
}

func (h *Orders) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:spot:orders:open", h.Open)
  mux.HandleFunc("binance:spot:orders:flush", h.Flush)
  mux.HandleFunc("binance:spot:orders:sync", h.Sync)
  return nil
}
