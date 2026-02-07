package futures

import (
  "context"
  "encoding/json"
  "fmt"
  "sync"
  "time"

  "github.com/hibiken/asynq"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  indicatorsRepositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type Indicators struct {
  AnsqContext *common.AnsqServerContext
  Repository  *repositories.IndicatorsRepository
}

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

type AndeanOscillatorPayload struct {
  Symbol   string
  Interval string
  Period   int
  Length   int
  Limit    int
}

type SuperTrendPayload struct {
	Symbol     string
	Interval   string
	Period     int
	Multiplier float64
	Limit      int
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
  baseRepository := indicatorsRepositories.BaseRepository{
    Db:  h.AnsqContext.Db,
    Rdb: h.AnsqContext.Rdb,
    Ctx: h.AnsqContext.Ctx,
  }
  h.Repository.Atr = &indicatorsRepositories.AtrRepository{BaseRepository: baseRepository}
  h.Repository.Pivot = &indicatorsRepositories.PivotRepository{BaseRepository: baseRepository}
  h.Repository.Kdj = &indicatorsRepositories.KdjRepository{BaseRepository: baseRepository}
  h.Repository.StochRsi = &indicatorsRepositories.StochRsiRepository{BaseRepository: baseRepository}
  h.Repository.Zlema = &indicatorsRepositories.ZlemaRepository{BaseRepository: baseRepository}
  h.Repository.HaZlema = &indicatorsRepositories.HaZlemaRepository{BaseRepository: baseRepository}
  h.Repository.BBands = &indicatorsRepositories.BBandsRepository{BaseRepository: baseRepository}
  h.Repository.AndeanOscillator = &indicatorsRepositories.AndeanOscillatorRepository{BaseRepository: baseRepository}
  h.Repository.IchimokuCloud = &indicatorsRepositories.IchimokuCloudRepository{BaseRepository: baseRepository}
  h.Repository.SuperTrend = &indicatorsRepositories.SuperTrendRepository{BaseRepository: baseRepository}
  h.Repository.VolumeProfile = &indicatorsRepositories.VolumeProfileRepository{BaseRepository: baseRepository}
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
    fmt.Sprintf("locks:binance:futures:indicators:pivot:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Pivot.Flush(payload.Symbol, payload.Interval)

  return nil
}

func (h *Indicators) Atr(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:indicators:atr:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Atr.Flush(payload.Symbol, payload.Interval, payload.Period, payload.Limit)

  return nil
}

func (h *Indicators) SuperTrend(ctx context.Context, t *asynq.Task) error {
	var payload SuperTrendPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
    return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
  }

  var mutex sync.Mutex
  mutex.Lock()
  defer mutex.Unlock()

  h.Repository.SuperTrend.Flush(payload.Symbol, payload.Interval, payload.Period, payload.Multiplier, payload.Limit)

  return nil
}

func (h *Indicators) Zlema(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:indicators:zlema:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Zlema.Flush(payload.Symbol, payload.Interval, payload.Period, payload.Limit)

  return nil
}

func (h *Indicators) HaZlema(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:indicators:ha_zlema:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.HaZlema.Flush(payload.Symbol, payload.Interval, payload.Period, payload.Limit)

  return nil
}

func (h *Indicators) Kdj(ctx context.Context, t *asynq.Task) error {
  var payload KdjPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:indicators:kdj:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.Kdj.Flush(payload.Symbol, payload.Interval, payload.LongPeriod, payload.ShortPeriod, payload.Limit)

  return nil
}

func (h *Indicators) BBands(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:indicators:bbands:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.BBands.Flush(payload.Symbol, payload.Interval, payload.Period, payload.Limit)

  return nil
}

func (h *Indicators) IchimokuCloud(ctx context.Context, t *asynq.Task) error {
  var payload IndicatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:indicators:ichimoku_cloud:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  if payload.Interval == "1m" {
    h.Repository.IchimokuCloud.Flush(payload.Symbol, payload.Interval, 129, 374, 748, 1440)
  } else if payload.Interval == "15m" {
    h.Repository.IchimokuCloud.Flush(payload.Symbol, payload.Interval, 60, 174, 349, 672)
  } else if payload.Interval == "4h" {
    h.Repository.IchimokuCloud.Flush(payload.Symbol, payload.Interval, 11, 32, 65, 126)
  } else {
    h.Repository.IchimokuCloud.Flush(payload.Symbol, payload.Interval, 9, 26, 52, 100)
  }

  return nil
}

func (h *Indicators) VolumeProfile(ctx context.Context, t *asynq.Task) error {
  var payload VolumeProfilePayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:indicators:volume_profile:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.VolumeProfile.Flush(payload.Symbol, payload.Interval, payload.Limit)

  return nil
}

func (h *Indicators) AndeanOscillator(ctx context.Context, t *asynq.Task) error {
  var payload AndeanOscillatorPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.AnsqContext.Rdb,
    h.AnsqContext.Ctx,
    fmt.Sprintf("locks:binance:futures:indicators:andean_oscillator:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  h.Repository.AndeanOscillator.Flush(payload.Symbol, payload.Interval, payload.Period, payload.Length, payload.Limit)

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
