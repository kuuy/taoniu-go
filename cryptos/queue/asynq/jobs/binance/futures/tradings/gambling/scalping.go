package gambling

import (
  "encoding/json"
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/futures"
)

type Scalping struct{}

func (h *Scalping) Place(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(ScalpingPlacePayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TRADINGS_GAMBLING_SCALPING_PLACE, payload), nil
}
