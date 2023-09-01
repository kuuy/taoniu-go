package futures

import (
  "github.com/hibiken/asynq"
)

type Tickers struct{}

func (h *Tickers) Flush() (*asynq.Task, error) {
  return asynq.NewTask("binance:futures:tickers:flush", nil), nil
}
