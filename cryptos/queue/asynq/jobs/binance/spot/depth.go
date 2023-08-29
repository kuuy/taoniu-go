package spot

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Depth struct{}

type DepthFlushPayload struct {
  Symbol   string
  Limit    int
  UseProxy bool
}

func (h *Depth) Flush(symbol string, limit int, useProxy bool) (*asynq.Task, error) {
  payload, err := json.Marshal(DepthFlushPayload{symbol, limit, useProxy})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:depth:flush", payload), nil
}
