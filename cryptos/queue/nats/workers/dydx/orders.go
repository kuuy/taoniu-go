package dydx

import (
  "encoding/json"

  "github.com/nats-io/nats.go"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/dydx"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type Orders struct {
  NatsContext *common.NatsContext
  Repository  *repositories.OrdersRepository
}

func NewOrders(natsContext *common.NatsContext) *Orders {
  h := &Orders{
    NatsContext: natsContext,
  }
  h.Repository = &repositories.OrdersRepository{
    Db:  h.NatsContext.Db,
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  return h
}

type OrdersUpdatePayload struct {
  Symbol  string `json:"symbol"`
  OrderId string `json:"order_id"`
  Status  string `json:"status"`
}

func (h *Orders) Subscribe() error {
  h.NatsContext.Conn.Subscribe(config.NATS_ORDERS_UPDATE, h.Update)
  return nil
}

func (h *Orders) Update(m *nats.Msg) {
  var payload *OrdersUpdatePayload
  json.Unmarshal(m.Data, &payload)

  if payload.Status == "PENDING" || payload.Status == "OPEN" {
    return
  }

  h.Repository.Flush(payload.OrderId)
}
