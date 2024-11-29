package spot

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type Indicators struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.IndicatorsRepository
}

func NewIndicators(ansqContext *common.AnsqServerContext) *Indicators {
  h := &Indicators{
    AnsqContext: ansqContext,
  }
  h.Repository = &repositories.IndicatorsRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
    Db: h.AnsqContext.Db,
  }
  return h
}

func (h *Indicators) Pivot(ctx context.Context, t *asynq.Task) error {
  var payload PivotPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:pivot:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Pivot(payload.Symbol, payload.Interval)

  return nil
}

func (h *Indicators) Atr(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:atr:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Atr(payload.Symbol, payload.Interval, payload.Period, payload.Limit)

  return nil
}

func (h *Indicators) Zlema(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:zlema:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Zlema(payload.Symbol, payload.Interval, payload.Period, payload.Limit)

  return nil
}

func (h *Indicators) HaZlema(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:ha_zlema:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.HaZlema(payload.Symbol, payload.Interval, payload.Period, payload.Limit)

  return nil
}

func (h *Indicators) Kdj(ctx context.Context, t *asynq.Task) error {
  var payload KdjPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:kdj:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Kdj(payload.Symbol, payload.Interval, payload.LongPeriod, payload.ShortPeriod, payload.Limit)

  return nil
}

func (h *Indicators) BBands(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:bbands:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.BBands(payload.Symbol, payload.Interval, payload.Period, payload.Limit)

  return nil
}

func (h *Indicators) IchimokuCloud(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:ichimoku_cloud:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  if payload.Interval == "1m" {
    h.Repository.IchimokuCloud(payload.Symbol, payload.Interval, 129, 374, 748, 1440)
  } else if payload.Interval == "15m" {
    h.Repository.IchimokuCloud(payload.Symbol, payload.Interval, 60, 174, 349, 672)
  } else if payload.Interval == "4h" {
    h.Repository.IchimokuCloud(payload.Symbol, payload.Interval, 11, 32, 65, 126)
  } else {
    h.Repository.IchimokuCloud(payload.Symbol, payload.Interval, 9, 26, 52, 100)
  }

  return nil
}

func (h *Indicators) VolumeProfile(ctx context.Context, t *asynq.Task) error {
  var payload VolumeProfilePayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:volume_profile:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.VolumeProfile(payload.Symbol, payload.Interval, payload.Limit)

  return nil
}

func (h *Indicators) AndeanOscillator(ctx context.Context, t *asynq.Task) error {
  var payload AndeanOscillatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:spot:indicators:andean_oscillator:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.AndeanOscillator(payload.Symbol, payload.Interval, payload.Period, payload.Length, payload.Limit)

  return nil
}

func (h *Indicators) Register() error {
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_INDICATORS_ATR, h.Atr)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_INDICATORS_ZLEMA, h.Zlema)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_INDICATORS_HA_ZLEMA, h.HaZlema)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_INDICATORS_KDJ, h.Kdj)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_INDICATORS_BBANDS, h.BBands)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_INDICATORS_ICHIMOKU_CLOUD, h.IchimokuCloud)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_INDICATORS_PIVOT, h.Pivot)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_INDICATORS_VOLUME_PROFILE, h.VolumeProfile)
  h.AnsqContext.Mux.HandleFunc(config.ASYNQ_JOBS_INDICATORS_ANDEAN_OSCILLATOR, h.AndeanOscillator)
  return nil
}
