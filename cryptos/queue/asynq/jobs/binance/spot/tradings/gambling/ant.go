package gambling

import (
  "encoding/json"
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/spot"
)

type Ant struct{}

func (h *Ant) Place(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(AntPlacePayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TRADINGS_GAMBLING_ANT_PLACE, payload), nil
}

func (h *Ant) Flush(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(AntFlushPayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TRADINGS_GAMBLING_ANT_FLUSH, payload), nil
}
