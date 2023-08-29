package dydx

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "github.com/nats-io/nats.go"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
  "time"
)

type Orderbook struct {
  Rdb        *redis.Client
  Ctx        context.Context
  Nats       *nats.Conn
  Repository *repositories.OrderbookRepository
}

func NewOrderbook() *Orderbook {
  h := &Orderbook{
    Rdb:  common.NewRedis(),
    Ctx:  context.Background(),
    Nats: common.NewNats(),
  }
  h.Repository = &repositories.OrderbookRepository{
    Rdb: h.Rdb,
    Ctx: h.Ctx,
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
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:dydx:orderbook:%s:%t", payload.Symbol, payload.UseProxy),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.Symbol)

  return nil
}

func (h *Orderbook) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("dydx:orderbook:flush", h.Flush)
  return nil
}
