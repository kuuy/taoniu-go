package dydx

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Orderbook struct{}

type OrderbookFlushPayload struct {
  Symbol   string
  UseProxy bool
}

func (h *Orderbook) Flush(symbol string, useProxy bool) (*asynq.Task, error) {
  payload, err := json.Marshal(OrderbookFlushPayload{symbol, useProxy})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:orderbook:flush", payload), nil
}
