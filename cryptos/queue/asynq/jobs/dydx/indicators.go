package dydx

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Indicators struct{}

type IndicatorPayload struct {
  Symbol   string
  Interval string
  Period   int
  Limit    int
}

type PivotPayload struct {
  Symbol   string
  Interval string
}

type KdjPayload struct {
  Symbol      string
  Interval    string
  LongPeriod  int
  ShortPeriod int
  Limit       int
}

type VolumeProfilePayload struct {
  Symbol   string
  Interval string
  Limit    int
}

func (h *Indicators) Pivot(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(PivotPayload{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:indicators:pivot", payload), nil
}

func (h *Indicators) Atr(symbol string, interval string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, interval, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:indicators:atr", payload), nil
}

func (h *Indicators) Zlema(symbol string, interval string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, interval, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:indicators:zlema", payload), nil
}

func (h *Indicators) HaZlema(symbol string, interval string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, interval, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:indicators:ha_zlema", payload), nil
}

func (h *Indicators) Kdj(symbol string, interval string, longPeriod int, shortPeriod int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(KdjPayload{symbol, interval, longPeriod, shortPeriod, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:indicators:kdj", payload), nil
}

func (h *Indicators) BBands(symbol string, interval string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, interval, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:indicators:bbands", payload), nil
}

func (h *Indicators) VolumeProfile(symbol string, interval string, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(VolumeProfilePayload{symbol, interval, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("dydx:indicators:volume_profile", payload), nil
}
