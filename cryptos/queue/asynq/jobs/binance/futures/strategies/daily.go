package strategies

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Daily struct{}

func (h *Daily) Atr(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:strategies:1d:atr", payload), nil
}

func (h *Daily) Zlema(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:strategies:1d:zlema", payload), nil
}

func (h *Daily) HaZlema(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:strategies:1d:hazlema", payload), nil
}

func (h *Daily) Kdj(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:strategies:1d:kdj", payload), nil
}

func (h *Daily) BBands(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:futures:strategies:1d:bbands", payload), nil
}
