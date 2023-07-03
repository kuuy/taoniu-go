package plans

import (
  "github.com/hibiken/asynq"
)

type Daily struct{}

func (h *Daily) Flush() (*asynq.Task, error) {
  return asynq.NewTask("binance:futures:plans:1d:flush", nil), nil
}
