package tradings

import (
  "encoding/json"
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/spot"
)

type Scalping struct{}

func (h *Scalping) Place(planId string) (*asynq.Task, error) {
  payload, err := json.Marshal(ScalpingPlacePayload{planId})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TRADINGS_SCALPING_PLACE, payload), nil
}

func (h *Scalping) Flush(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(ScalpingFlushPayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TRADINGS_SCALPING_FLUSH, payload), nil
}
