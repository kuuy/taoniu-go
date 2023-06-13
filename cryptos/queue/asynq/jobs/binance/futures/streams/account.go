package streams

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Account struct{}

type AccountUpdatePayload struct {
  Balance   float64
  Timestamp int64
}

func (h *Account) Update(balance float64, timestamp int64) (*asynq.Task, error) {
  payload, err := json.Marshal(AccountUpdatePayload{balance, timestamp})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:account:update", payload), nil
}
