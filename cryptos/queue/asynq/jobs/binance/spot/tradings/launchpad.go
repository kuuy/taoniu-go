package tradings

import (
  "encoding/json"
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/spot"
)

type Launchpad struct{}

func (h *Launchpad) Place(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(LaunchpadPlacePayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TRADINGS_LAUNCHPAD_PLACE, payload), nil
}

func (h *Launchpad) Flush(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(LaunchpadFlushPayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TRADINGS_LAUNCHPAD_FLUSH, payload), nil
}
