package futures

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Plans struct{}

type PlansPayload struct {
  Interval string
}

func (h *Plans) Flush(interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(PlansPayload{interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:plans:flush", payload), nil
}
