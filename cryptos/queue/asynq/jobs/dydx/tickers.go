package dydx

import (
  "github.com/hibiken/asynq"
)

type Tickers struct{}

func (h *Tickers) Flush() (*asynq.Task, error) {
  return asynq.NewTask("dydx:tickers:flush", nil), nil
}
