package futures

import (
  "encoding/json"
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/futures"
)

type Klines struct{}

func (h *Klines) Flush(symbol string, interval string, limit int, useProxy bool) (*asynq.Task, error) {
  payload, err := json.Marshal(KlinesFlushPayload{symbol, interval, limit, useProxy})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_KLINES_FLUSH, payload), nil
}
