package tradings

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Scalping struct{}

type ScalpingPlacePayload struct {
  PlanID string `json:"plan_id"`
}

type ScalpingFlushPayload struct {
  ID string `json:"id"`
}

func (h *Scalping) Flush(planID string) (*asynq.Task, error) {
  payload, err := json.Marshal(ScalpingFlushPayload{planID})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:tradings:scalping:flush", payload), nil
}

func (h *Scalping) Place(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(ScalpingPlacePayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:tradings:scalping:place", payload), nil
}
