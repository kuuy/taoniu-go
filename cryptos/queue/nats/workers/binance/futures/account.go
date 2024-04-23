package futures

import (
  "encoding/json"
  "fmt"
  "github.com/nats-io/nats.go"
  "taoniu.local/cryptos/common"
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

func (h *Account) Subscribe() error {
  //h.NatsContext.Conn.Subscribe(config.NATS_ACCOUNT_UPDATE, h.Update)
  return nil
}

func (h *Account) Update(m *nats.Msg) {
  var payload *AccountUpdatePayload
  json.Unmarshal(m.Data, &payload)

  h.NatsContext.Rdb.HMSet(
    h.NatsContext.Ctx,
    fmt.Sprintf("binance:futures:balance:%s", payload.Asset),
    map[string]interface{}{
      "balance": payload.Balance,
    },
  )
}
