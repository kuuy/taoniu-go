package cross

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Positions struct{}

type PositionsFlushPayload struct {
  Symbol string `json:"symbol"`
  Side   int    `json:"side"`
}

func (h *Positions) Flush(symbol string, side int) (*asynq.Task, error) {
  payload, err := json.Marshal(PositionsFlushPayload{symbol, side})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:margin:cross:positions:flush", payload), nil
}
