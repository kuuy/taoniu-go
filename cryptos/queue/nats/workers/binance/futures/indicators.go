package futures

import (
  "encoding/json"
  "fmt"
  "time"

  "github.com/nats-io/nats.go"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  indicatorsRepositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type Indicators struct {
  NatsContext *common.NatsContext
  Repository  *repositories.IndicatorsRepository
}

func NewIndicators(natsContext *common.NatsContext) *Indicators {
  h := &Indicators{
    NatsContext: natsContext,
  }
  h.Repository = &repositories.IndicatorsRepository{
    Db:  h.NatsContext.Db,
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  baseRepository := indicatorsRepositories.BaseRepository{
    Db:  h.NatsContext.Db,
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  h.Repository.Atr = &indicatorsRepositories.AtrRepository{BaseRepository: baseRepository}
  h.Repository.BBands = &indicatorsRepositories.BBandsRepository{BaseRepository: baseRepository}
  h.Repository.Pivot = &indicatorsRepositories.PivotRepository{BaseRepository: baseRepository}
  h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
    Db: h.NatsContext.Db,
  }
  return h
}

func (h *Indicators) Subscribe() error {
  h.NatsContext.Conn.Subscribe(config.NATS_KLINES_UPDATE, h.Flush)
  return nil
}

func (h *Indicators) Pivot(symbol string, interval string) error {
  return h.Repository.Pivot.Flush(symbol, interval)
}

func (h *Indicators) Atr(symbol string, interval string) error {
  return h.Repository.Atr.Flush(symbol, interval, 14, 100)
}

func (h *Indicators) Zlema(symbol string, interval string) error {
  return h.Repository.Zlema(symbol, interval, 14, 100)
}

func (h *Indicators) HaZlema(symbol string, interval string) error {
  return h.Repository.HaZlema(symbol, interval, 14, 100)
}

func (h *Indicators) Kdj(symbol string, interval string) error {
  return h.Repository.Kdj(symbol, interval, 9, 3, 100)
}

func (h *Indicators) BBands(symbol string, interval string) error {
  return h.Repository.BBands.Flush(symbol, interval, 14, 100)
}

func (h *Indicators) IchimokuCloud(symbol string, interval string) error {
  if interval == "1m" {
    return h.Repository.IchimokuCloud(symbol, interval, 129, 374, 748, 1440)
  } else if interval == "15m" {
    return h.Repository.IchimokuCloud(symbol, interval, 60, 174, 349, 672)
  } else if interval == "4h" {
    return h.Repository.IchimokuCloud(symbol, interval, 11, 32, 65, 126)
  } else {
    return h.Repository.IchimokuCloud(symbol, interval, 9, 26, 52, 100)
  }
}

func (h *Indicators) VolumeProfile(symbol string, interval string) error {
  var limit int
  if interval == "1m" {
    limit = 1440
  } else if interval == "15m" {
    limit = 672
  } else if interval == "4h" {
    limit = 126
  } else {
    limit = 100
  }
  return h.Repository.VolumeProfile(symbol, interval, limit)
}

func (h *Indicators) AndeanOscillator(symbol string, interval string, period int, length int) error {
  var limit int
  if interval == "1m" {
    limit = 1440
  } else if interval == "15m" {
    limit = 672
  } else if interval == "4h" {
    limit = 126
  } else {
    limit = 100
  }
  return h.Repository.AndeanOscillator(symbol, interval, period, length, limit)
}

func (h *Indicators) Flush(m *nats.Msg) {
  var payload *KlinesUpdatePayload
  json.Unmarshal(m.Data, &payload)

  mutex := common.NewMutex(
    h.NatsContext.Rdb,
    h.NatsContext.Ctx,
    fmt.Sprintf(config.LOCKS_INDICATORS_FLUSH, payload.Interval, payload.Symbol),
  )
  if !mutex.Lock(30 * time.Second) {
    return
  }
  defer mutex.Unlock()

  h.Pivot(payload.Symbol, payload.Interval)
  h.Atr(payload.Symbol, payload.Interval)
  h.Zlema(payload.Symbol, payload.Interval)
  h.HaZlema(payload.Symbol, payload.Interval)
  h.Kdj(payload.Symbol, payload.Interval)
  h.BBands(payload.Symbol, payload.Interval)
  h.IchimokuCloud(payload.Symbol, payload.Interval)
  h.VolumeProfile(payload.Symbol, payload.Interval)
  h.AndeanOscillator(payload.Symbol, payload.Interval, 50, 9)

  h.NatsContext.Conn.Publish(config.NATS_INDICATORS_UPDATE, m.Data)
  h.NatsContext.Conn.Flush()
}
