package futures

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Tickers struct{}

type TickersFlushPayload struct {
  Symbol   string
  UseProxy bool
}

func (h *Tickers) Flush(symbol string, useProxy bool) (*asynq.Task, error) {
  payload, err := json.Marshal(TickersFlushPayload{symbol, useProxy})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:tickers:flush", payload), nil
}
