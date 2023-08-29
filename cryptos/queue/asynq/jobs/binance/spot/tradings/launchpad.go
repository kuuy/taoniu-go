package tradings

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Launchpad struct{}

type LaunchpadFlushPayload struct {
  ID string
}

type LaunchpadPlacePayload struct {
  ID string
}

func (h *Launchpad) Flush(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(LaunchpadFlushPayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:tradings:launchpad:flush", payload), nil
}

func (h *Launchpad) Place(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(LaunchpadPlacePayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:tradings:launchpad:place", payload), nil
}
