package futures

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "gorm.io/gorm"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
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
    Db: h.Db,
  }
  return h
}

type OrdersOpenPayload struct {
  Symbol string
}

type OrdersFlushPayload struct {
  Symbol  string
  OrderID int64
}

func (h *Orders) Open(ctx context.Context, t *asynq.Task) error {
  var payload OrdersFlushPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:binance:futures:tradings:orders:open:%s", payload.Symbol),
  )
  if mutex.Lock(5 * time.Second) {
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
    fmt.Sprintf("locks:binance:futures:tradings:orders:flush:%s:%d", payload.Symbol, payload.OrderID),
  )
  if mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.Symbol, payload.OrderID)

  return nil
}

func (h *Orders) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:futures:orders:open", h.Open)
  mux.HandleFunc("binance:futures:orders:flush", h.Flush)
  return nil
}
