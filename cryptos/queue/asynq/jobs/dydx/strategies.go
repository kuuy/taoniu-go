package dydx

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Strategies struct{}

type StrategyPayload struct {
  Symbol   string
  Interval string
}

func (h *Strategies) Atr(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:strategies:atr", payload), nil
}

func (h *Strategies) Zlema(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:strategies:zlema", payload), nil
}

func (h *Strategies) HaZlema(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:strategies:ha_zlema", payload), nil
}

func (h *Strategies) Kdj(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:strategies:kdj", payload), nil
}

func (h *Strategies) BBands(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:strategies:bbands", payload), nil
}
