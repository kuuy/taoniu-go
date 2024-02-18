package streams

import (
  "encoding/json"
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/futures"
)

type Klines struct{}

type KlinesUpdatePayload struct {
  Symbol    string
  Interval  string
  Open      float64
  Close     float64
  High      float64
  Low       float64
  Volume    float64
  Quota     float64
  Timestamp int64
}

func (h *Klines) Update(
  symbol string,
  interval string,
  open float64,
  close float64,
  high float64,
  low float64,
  volume float64,
  quota float64,
  timestamp int64,
) (*asynq.Task, error) {
  payload, err := json.Marshal(KlinesUpdatePayload{
    symbol,
    interval,
    open,
    close,
    high,
    low,
    volume,
    quota,
    timestamp,
  })
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_KLINES_UPDATE, payload), nil
}
