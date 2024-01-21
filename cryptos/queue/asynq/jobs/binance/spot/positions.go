package spot

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Positions struct{}

type PositionsFlushPayload struct {
  Symbol string `json:"symbol"`
}

func (h *Positions) Flush(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(PositionsFlushPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:positions:flush", payload), nil
}
