package dydx

import (
  "encoding/json"
  "fmt"
  "taoniu.local/cryptos/common"
  "time"

  "github.com/nats-io/nats.go"

  config "taoniu.local/cryptos/config/dydx"
)

type Trades struct {
  NatsContext *common.NatsContext
}

func NewTrades(natsContext *common.NatsContext) *Trades {
  h := &Trades{
    NatsContext: natsContext,
  }
  return h
}

type TradesUpdatePayload struct {
  Symbol string  `json:"symbol"`
  Price  float64 `json:"price"`
  Side   string  `json:"side"`
}

func (h *Trades) Subscribe() error {
  h.NatsContext.Conn.Subscribe(config.NATS_TRADES_UPDATE, h.Update)
  return nil
}

func (h *Trades) Update(m *nats.Msg) {
  var payload *TradesUpdatePayload
  json.Unmarshal(m.Data, &payload)

  data := map[string]interface{}{}
  if payload.Side == "BUY" {
    data["bid"] = payload.Price
  }
  if payload.Side == "SELL" {
    data["ask"] = payload.Price
  }
  data["timestamp"] = time.Now().UnixMilli()

  h.NatsContext.Rdb.HMSet(
    h.NatsContext.Ctx,
    fmt.Sprintf("dydx.prices:%s", payload.Symbol),
    data,
  )
}
