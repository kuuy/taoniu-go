package spot

import (
  "encoding/json"
  "github.com/hibiken/asynq"
  config "taoniu.local/cryptos/config/binance/spot"
)

type Tickers struct{}

func (h *Tickers) Flush(symbols []string, useProxy bool) (*asynq.Task, error) {
  payload, err := json.Marshal(TickersFlushPayload{symbols, useProxy})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TICKERS_FLUSH, payload), nil
}

func (h *Tickers) Update(
  symbol string,
  price float64,
  open float64,
  hign float64,
  low float64,
  volume float64,
  quota float64,
  timestamp int64,
) (*asynq.Task, error) {
  payload, err := json.Marshal(TickersUpdatePayload{
    symbol,
    price,
    open,
    hign,
    low,
    volume,
    quota,
    timestamp,
  })
  if err != nil {
    return nil, err
  }
  return asynq.NewTask(config.ASYNQ_JOBS_TICKERS_UPDATE, payload), nil
}
