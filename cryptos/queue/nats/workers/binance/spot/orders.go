package spot

import (
  "encoding/json"

  "github.com/nats-io/nats.go"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
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
    Ctx: h.NatsContext.Ctx,
  }
  return h
}

type OrdersUpdatePayload struct {
  Symbol  string `json:"symbol"`
  OrderId int64  `json:"order_id"`
  Status  string `json:"status"`
}

func (h *Orders) Subscribe() error {
  h.NatsContext.Conn.Subscribe(config.NATS_ORDERS_UPDATE, h.Update)
  return nil
}

func (h *Orders) Update(m *nats.Msg) {
  var payload *OrdersUpdatePayload
  json.Unmarshal(m.Data, &payload)

  if payload.Status == "NEW" || payload.Status == "PARTIALLY_FILLED" {
    return
  }

  h.Repository.Flush(payload.Symbol, payload.OrderId)
}
