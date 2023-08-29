package dydx

import (
  "github.com/hibiken/asynq"
)

type Account struct{}

func (h *Account) Flush() (*asynq.Task, error) {
  return asynq.NewTask("dydx:account:flush", nil), nil
}
