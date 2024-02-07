package futures

import (
  "encoding/json"
  "errors"
  "github.com/nats-io/nats.go"
  socketio "github.com/vchitai/go-socket.io/v4"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
)

type Tickers struct {
  SocketContext  *common.SocketContext
  FlushSubscribe *nats.Subscription
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

func NewTickers(socketContext *common.SocketContext) *Tickers {
  return &Tickers{
    SocketContext: socketContext,
  }
}

func (h *Tickers) Flush(symbols []string) error {
  if h.FlushSubscribe != nil && h.FlushSubscribe.IsValid() {
    h.FlushSubscribe.Unsubscribe()
  }
  h.FlushSubscribe, _ = h.SocketContext.Nats.Subscribe(config.NATS_TICKERS_UPDATE, func(msg *nats.Msg) {
    var payload *TickersUpdatePayload
    json.Unmarshal(msg.Data, &payload)
    if h.contains(symbols, payload.Symbol) {
      h.SocketContext.Conn.Emit("tickers", payload)
    }
  })
  return nil
}

func (h *Tickers) Register(s *socketio.Server) error {
  return nil
}

func (h *Tickers) Subscribe(req map[string]interface{}) error {
  if _, ok := req["symbols"]; !ok {
    return errors.New("symbols is empty")
  }
  var symbols []string
  for _, symbol := range req["symbols"].([]interface{}) {
    symbols = append(symbols, symbol.(string))
  }
  return h.Flush(symbols)
}

func (h *Tickers) UnSubscribe() error {
  if h.FlushSubscribe != nil {
    h.FlushSubscribe.Unsubscribe()
  }
  return nil
}

func (h *Tickers) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
