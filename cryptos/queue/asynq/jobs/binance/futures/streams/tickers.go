package streams

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Tickers struct{}

type TickersUpdatePayload struct {
  Symbol    string
  Open      float64
  Price     float64
  High      float64
  Low       float64
  Volume    float64
  Quota     float64
  Timestamp int64
}

func (h *Tickers) Update(
  symbol string,
  open float64,
  price float64,
  high float64,
  low float64,
  volume float64,
  quota float64,
  timestamp int64,
) (*asynq.Task, error) {
  payload, err := json.Marshal(TickersUpdatePayload{
    symbol,
    open,
    price,
    high,
    low,
    volume,
    quota,
    timestamp,
  })
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:tickers:update", payload), nil
}
