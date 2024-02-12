package spot

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
  return asynq.NewTask("binance:spot:strategies:atr", payload), nil
}

func (h *Strategies) Zlema(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:strategies:zlema", payload), nil
}

func (h *Strategies) HaZlema(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:strategies:ha_zlema", payload), nil
}

func (h *Strategies) Kdj(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:strategies:kdj", payload), nil
}

func (h *Strategies) BBands(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:strategies:bbands", payload), nil
}

func (h *Strategies) IchimokuCloud(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(StrategyPayload{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:strategies:ichimoku_cloud", payload), nil
}
