package futures

import (
  "context"
  "encoding/json"
  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "gorm.io/gorm"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type Orders struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.OrdersRepository
}

func NewOrders(db *gorm.DB, rdb *redis.Client, ctx context.Context) *Orders {
  h := &Orders{
    Db:  db,
    Rdb: rdb,
    Ctx: ctx,
  }
  h.Repository = &repositories.OrdersRepository{
    Db:  h.Db,
    Ctx: h.Ctx,
  }
  return h
}

type OrdersUpdatePayload struct {
  Symbol  string `json:"symbol"`
  OrderID int64  `json:"order_id"`
  Status  string `json:"status"`
}

func (h *Orders) Subscribe(nc *nats.Conn) error {
  nc.Subscribe(config.NATS_ORDERS_UPDATE, h.Update)
  return nil
}

func (h *Orders) Update(m *nats.Msg) {
  var payload *OrdersUpdatePayload
  json.Unmarshal(m.Data, &payload)

  if payload.Status == "NEW" || payload.Status == "PARTIALLY_FILLED" {
    return
  }

  h.Repository.Flush(payload.Symbol, payload.OrderID)
}
