package spot

import (
  "encoding/json"
  "fmt"
  "github.com/nats-io/nats.go"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
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
  Asset  string  `json:"asset"`
  Free   float64 `json:"free"`
  Locked float64 `json:"locked"`
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
    fmt.Sprintf("binance:spot:balance:%s", payload.Asset),
    map[string]interface{}{
      "free":   payload.Free,
      "locked": payload.Locked,
    },
  )
}
