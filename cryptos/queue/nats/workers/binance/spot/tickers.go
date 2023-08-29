package spot

import (
  "encoding/json"
  "fmt"
  "github.com/nats-io/nats.go"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
)

type Tickers struct {
  NatsContext *common.NatsContext
}

func NewTickers(natsContext *common.NatsContext) *Tickers {
  h := &Tickers{
    NatsContext: natsContext,
  }
  return h
}

type TickersUpdatePayload struct {
  Symbol    string  `json:"symbol"`
  Open      float64 `json:"open"`
  Price     float64 `json:"price"`
  High      float64 `json:"high"`
  Low       float64 `json:"low"`
  Volume    float64 `json:"volume"`
  Quota     float64 `json:"quota"`
  Timestamp int64   `json:"timestamp"`
}

func (h *Tickers) Subscribe() error {
  h.NatsContext.Conn.Subscribe(config.NATS_TRADES_UPDATE, h.Update)
  return nil
}

func (h *Tickers) Update(m *nats.Msg) {
  var payload *TickersUpdatePayload
  json.Unmarshal(m.Data, &payload)

  h.NatsContext.Rdb.HMSet(
    h.NatsContext.Ctx,
    fmt.Sprintf("binance:spot:realtime:%s", payload.Symbol),
    map[string]interface{}{
      "symbol":    payload.Symbol,
      "price":     payload.Price,
      "open":      payload.Open,
      "high":      payload.High,
      "low":       payload.Low,
      "volume":    payload.Volume,
      "quota":     payload.Quota,
      "timestamp": payload.Timestamp,
    },
  )
}
