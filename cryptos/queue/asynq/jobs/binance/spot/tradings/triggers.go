package tradings

import (
  "encoding/json"
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/spot"
)

type Triggers struct{}

func (h *Triggers) Place(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(TriggersPlacePayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TRADINGS_TRIGGERS_PLACE, payload), nil
}

func (h *Triggers) Flush(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(TriggersFlushPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TRADINGS_TRIGGERS_FLUSH, payload), nil
}
