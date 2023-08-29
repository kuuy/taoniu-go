package spot

import (
  "github.com/hibiken/asynq"
)

type Account struct{}

func (h *Account) Flush() (*asynq.Task, error) {
  return asynq.NewTask("binance:spot:account:flush", nil), nil
}
