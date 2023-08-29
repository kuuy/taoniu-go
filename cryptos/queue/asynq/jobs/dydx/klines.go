package dydx

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Klines struct{}

type KlinesFlushPayload struct {
  Symbol   string
  Interval string
  Endtime  int64
  Limit    int
  UseProxy bool
}

func (h *Klines) Flush(symbol string, interval string, endtime int64, limit int, useProxy bool) (*asynq.Task, error) {
  payload, err := json.Marshal(KlinesFlushPayload{symbol, interval, endtime, limit, useProxy})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:klines:flush", payload), nil
}
