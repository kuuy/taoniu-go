package tradings

import (
  "encoding/json"
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/spot"
)

type Triggers struct{}

func (h *Triggers) Place(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(TriggersPlacePayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TRADINGS_TRIGGERS_PLACE, payload), nil
}

func (h *Triggers) Flush(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(TriggersFlushPayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TRADINGS_TRIGGERS_FLUSH, payload), nil
}
