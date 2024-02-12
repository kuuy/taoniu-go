package spot

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

type IchimokuCloud struct {
  Symbol   string
  Interval string
}

type VolumeProfilePayload struct {
  Symbol   string
  Interval string
  Limit    int
}

type AndeanOscillatorPayload struct {
  Symbol   string
  Interval string
  Period   int
  Length   int
  Limit    int
}

func (h *Indicators) Pivot(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(PivotPayload{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:pivot", payload), nil
}

func (h *Indicators) Atr(symbol string, interval string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, interval, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:atr", payload), nil
}

func (h *Indicators) Zlema(symbol string, interval string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, interval, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:zlema", payload), nil
}

func (h *Indicators) HaZlema(symbol string, interval string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, interval, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:ha_zlema", payload), nil
}

func (h *Indicators) Kdj(symbol string, interval string, longPeriod int, shortPeriod int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(KdjPayload{symbol, interval, longPeriod, shortPeriod, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:kdj", payload), nil
}

func (h *Indicators) BBands(symbol string, interval string, period int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(IndicatorPayload{symbol, interval, period, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:bbands", payload), nil
}

func (h *Indicators) IchimokuCloud(symbol string, interval string) (*asynq.Task, error) {
  payload, err := json.Marshal(IchimokuCloud{symbol, interval})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:ichimoku_cloud", payload), nil
}

func (h *Indicators) VolumeProfile(symbol string, interval string, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(VolumeProfilePayload{symbol, interval, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:volume_profile", payload), nil
}

func (h *Indicators) AndeanOscillator(symbol string, interval string, period int, length int, limit int) (*asynq.Task, error) {
  payload, err := json.Marshal(AndeanOscillatorPayload{symbol, interval, period, length, limit})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("binance:spot:indicators:andean_oscillator", payload), nil
}
