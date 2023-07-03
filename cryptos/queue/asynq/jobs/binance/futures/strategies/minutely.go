package strategies

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Minutely struct{}

func (h *Minutely) Atr(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:strategies:1m:atr", payload), nil
}

func (h *Minutely) Zlema(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:strategies:1m:zlema", payload), nil
}

func (h *Minutely) HaZlema(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:strategies:1m:hazlema", payload), nil
}

func (h *Minutely) Kdj(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:strategies:1m:kdj", payload), nil
}

func (h *Minutely) BBands(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:strategies:1m:bbands", payload), nil
}
