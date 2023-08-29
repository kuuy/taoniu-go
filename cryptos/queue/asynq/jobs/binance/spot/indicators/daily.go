package indicators

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Daily struct{}

func (h *Daily) Pivot(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(PivotPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1d:pivot", payload), nil
}

func (h *Daily) Atr(symbol string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1d:atr", payload), nil
}

func (h *Daily) Zlema(symbol string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1d:zlema", payload), nil
}

func (h *Daily) HaZlema(symbol string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1d:hazlema", payload), nil
}

func (h *Daily) Kdj(symbol string, longPeriod int, shortPeriod int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(KdjPayload{symbol, longPeriod, shortPeriod, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1d:kdj", payload), nil
}

func (h *Daily) BBands(symbol string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1d:bbands", payload), nil
}

func (h *Daily) VolumeProfile(symbol string, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(VolumeProfilePayload{symbol, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1d:volume_profile", payload), nil
}
