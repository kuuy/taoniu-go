package futures

import (
  "encoding/json"
  "github.com/nats-io/nats.go"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type Strategies struct {
  NatsContext *common.NatsContext
  Repository  *repositories.StrategiesRepository
}

func NewStrategies(natsContext *common.NatsContext) *Strategies {
  h := &Strategies{
    NatsContext: natsContext,
  }
  h.Repository = &repositories.StrategiesRepository{
    Db:  h.NatsContext.Db,
    Rdb: h.NatsContext.Rdb,
    Ctx: h.NatsContext.Ctx,
  }
  h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
    Db: h.NatsContext.Db,
  }
  return h
}

func (h *Strategies) Subscribe() error {
  h.NatsContext.Conn.Subscribe(config.NATS_INDICATORS_UPDATE, h.Flush)
  return nil
}

func (h *Strategies) Atr(symbol string, interval string) error {
  return h.Repository.Atr(symbol, interval)
}

func (h *Strategies) Zlema(symbol string, interval string) error {
  return h.Repository.Zlema(symbol, interval)
}

func (h *Strategies) HaZlema(symbol string, interval string) error {
  return h.Repository.HaZlema(symbol, interval)
}

func (h *Strategies) Kdj(symbol string, interval string) error {
  return h.Repository.Kdj(symbol, interval)
}

func (h *Strategies) BBands(symbol string, interval string) error {
  return h.Repository.BBands(symbol, interval)
}

func (h *Strategies) Flush(m *nats.Msg) {
  var payload *IndicatorsUpdatePayload
  json.Unmarshal(m.Data, &payload)

  h.Atr(payload.Symbol, payload.Interval)
  h.Zlema(payload.Symbol, payload.Interval)
  h.HaZlema(payload.Symbol, payload.Interval)
  h.Kdj(payload.Symbol, payload.Interval)
  h.BBands(payload.Symbol, payload.Interval)

  h.NatsContext.Conn.Publish(config.NATS_STRATEGIES_UPDATE, m.Data)
  h.NatsContext.Conn.Flush()
}
