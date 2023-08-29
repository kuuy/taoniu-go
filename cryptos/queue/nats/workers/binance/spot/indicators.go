package spot

import (
  "encoding/json"
  "github.com/nats-io/nats.go"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
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
  return h.Repository.Pivot(symbol, interval)
}

func (h *Indicators) Atr(symbol string, interval string) error {
  return h.Repository.Atr(symbol, interval, 14, 100)
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
  return h.Repository.BBands(symbol, interval, 14, 100)
}

func (h *Indicators) VolumeProfile(symbol string, interval string) error {
  var limit int
  if interval == "1m" {
    limit = 1440
  } else if interval == "4h" {
    limit = 126
  } else {
    limit = 100
  }
  return h.Repository.VolumeProfile(symbol, interval, limit)
}

func (h *Indicators) Flush(m *nats.Msg) {
  var payload *KlinesUpdatePayload
  json.Unmarshal(m.Data, &payload)

  h.Pivot(payload.Symbol, payload.Interval)
  h.Atr(payload.Symbol, payload.Interval)
  h.Zlema(payload.Symbol, payload.Interval)
  h.HaZlema(payload.Symbol, payload.Interval)
  h.Kdj(payload.Symbol, payload.Interval)
  h.BBands(payload.Symbol, payload.Interval)
  h.VolumeProfile(payload.Symbol, payload.Interval)

  h.NatsContext.Conn.Publish(config.NATS_INDICATORS_UPDATE, m.Data)
  h.NatsContext.Conn.Flush()
}
