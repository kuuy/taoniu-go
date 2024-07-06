package cross

import (
  "encoding/json"
  "fmt"
  "github.com/nats-io/nats.go"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/margin/cross"
)

type Account struct {
  NatsContext *common.NatsContext
}

func NewAccount(natsContext *common.NatsContext) *Account {
  h := &Account{
    NatsContext: natsContext,
  }
  return h
}

type AccountUpdatePayload struct {
  Asset    string  `json:"asset"`
  Free     float64 `json:"free"`
  Locked   float64 `json:"locked"`
  Borrowed float64 `json:"borrowed"`
  Interest float64 `json:"interest"`
}

func (h *Account) Subscribe() error {
  h.NatsContext.Conn.Subscribe(config.NATS_ACCOUNT_UPDATE, h.Update)
  return nil
}

func (h *Account) Update(m *nats.Msg) {
  var payload *AccountUpdatePayload
  json.Unmarshal(m.Data, &payload)

  h.NatsContext.Rdb.HMSet(
    h.NatsContext.Ctx,
    fmt.Sprintf("binance:margin:cross:balance:%s", payload.Asset),
    map[string]interface{}{
      "free":     payload.Free,
      "locked":   payload.Locked,
      "borrowed": payload.Borrowed,
      "interest": payload.Interest,
    },
  )
}
