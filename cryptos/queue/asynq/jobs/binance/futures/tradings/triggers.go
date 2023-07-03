package tradings

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Triggers struct{}

type TriggersFlushPayload struct {
  ID string
}

type TriggersPlacePayload struct {
  ID string
}

func (h *Triggers) Flush(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(TriggersFlushPayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:tradings:triggers:flush", payload), nil
}

func (h *Triggers) Place(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(TriggersPlacePayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:tradings:triggers:place", payload), nil
}
