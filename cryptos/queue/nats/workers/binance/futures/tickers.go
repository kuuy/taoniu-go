package futures

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  config "taoniu.local/cryptos/config/binance/futures"
)

type Tickers struct {
  Rdb *redis.Client
  Ctx context.Context
}

func NewTickers(rdb *redis.Client, ctx context.Context) *Tickers {
  h := &Tickers{
    Rdb: rdb,
    Ctx: ctx,
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

func (h *Tickers) Subscribe(nc *nats.Conn) error {
  nc.Subscribe(config.NATS_TICKERS_UPDATE, h.Update)
  return nil
}

func (h *Tickers) Update(m *nats.Msg) {
  var payload *TickersUpdatePayload
  json.Unmarshal(m.Data, &payload)

  h.Rdb.HMSet(
    h.Ctx,
    fmt.Sprintf("binance:futures:realtime:%s", payload.Symbol),
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
