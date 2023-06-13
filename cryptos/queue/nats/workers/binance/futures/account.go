package futures

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  config "taoniu.local/cryptos/config/binance/futures"
)

type Account struct {
  Rdb *redis.Client
  Ctx context.Context
}

func NewAccount(rdb *redis.Client, ctx context.Context) *Account {
  h := &Account{
    Rdb: rdb,
    Ctx: ctx,
  }
  return h
}

type AccountUpdatePayload struct {
  Asset            string  `json:"asset"`
  Balance          float64 `json:"balance"`
  AvailableBalance float64 `json:"available_balance"`
}

func (h *Account) Subscribe(nc *nats.Conn) error {
  nc.Subscribe(config.NATS_ACCOUNT_UPDATE, h.Update)
  return nil
}

func (h *Account) Update(m *nats.Msg) {
  var payload *AccountUpdatePayload
  json.Unmarshal(m.Data, &payload)

  h.Rdb.HMSet(
    h.Ctx,
    fmt.Sprintf("binance:futures:balance:%s", payload.Asset),
    map[string]interface{}{
      "balance":           payload.Balance,
      "available_balance": payload.AvailableBalance,
    },
  )
}
