package futures

import (
  "encoding/json"
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/futures"
)

type FundingRate struct{}

func (h *FundingRate) Flush() (*asynq.Task, error) {
  payload, err := json.Marshal(FundingRateFlushPayload{})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_FUNDING_RATE_FLUSH, payload), nil
}
