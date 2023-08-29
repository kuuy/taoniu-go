package futures

import (
  "github.com/hibiken/asynq"
)

type Account struct{}

func (h *Account) Flush() (*asynq.Task, error) {
  return asynq.NewTask("binance:futures:account:flush", nil), nil
}
