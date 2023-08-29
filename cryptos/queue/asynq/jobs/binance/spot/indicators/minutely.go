package indicators

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Minutely struct{}

func (h *Minutely) Pivot(symbol string) (*asynq.Task, error) {
  payload, err := json.Marshal(PivotPayload{symbol})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1m:pivot", payload), nil
}

func (h *Minutely) Atr(symbol string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1m:atr", payload), nil
}

func (h *Minutely) Zlema(symbol string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1m:zlema", payload), nil
}

func (h *Minutely) HaZlema(symbol string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1m:hazlema", payload), nil
}

func (h *Minutely) Kdj(symbol string, longPeriod int, shortPeriod int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(KdjPayload{symbol, longPeriod, shortPeriod, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1m:kdj", payload), nil
}

func (h *Minutely) BBands(symbol string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1m:bbands", payload), nil
}

func (h *Minutely) VolumeProfile(symbol string, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(VolumeProfilePayload{symbol, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:1m:volume_profile", payload), nil
}
