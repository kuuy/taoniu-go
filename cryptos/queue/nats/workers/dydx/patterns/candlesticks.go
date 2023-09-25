package patterns

import (
  "encoding/json"
  "fmt"
  "time"

  "github.com/nats-io/nats.go"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/dydx"
  repositories "taoniu.local/cryptos/repositories/dydx/patterns"
)

type Candlesticks struct {
  NatsContext *common.NatsContext
  Repository  *repositories.CandlesticksRepository
}

func NewCandlesticks(natsContext *common.NatsContext) *Candlesticks {
  h := &Candlesticks{
    NatsContext: natsContext,
  }
  h.Repository = &repositories.CandlesticksRepository{
    Db: h.NatsContext.Db,
  }
  return h
}

func (h *Candlesticks) Subscribe() error {
  h.NatsContext.Conn.Subscribe(config.NATS_KLINES_UPDATE, h.Flush)
  return nil
}

func (h *Candlesticks) Flush(m *nats.Msg) {
  var payload *KlinesUpdatePayload
  json.Unmarshal(m.Data, &payload)

  mutex := common.NewMutex(
    h.NatsContext.Rdb,
    h.NatsContext.Ctx,
    fmt.Sprintf("locks:dydx:patterns:candlesticks:%s:%s", payload.Symbol, payload.Interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return
  }
  defer mutex.Unlock()

  h.Repository.Flush(payload.Symbol, payload.Interval, 20)
}
