package spot

import (
  "encoding/json"
  "errors"
  "fmt"
  "strconv"
  "strings"
  "sync"

  "github.com/nats-io/nats.go"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
)

type HubSubscriber interface {
  Subscribe(client interface{}, topic string)
  Unsubscribe(client interface{}, topic string)
  Broadcast(topic string, message []byte)
}

type ClientSender interface {
  SendBytes(data []byte)
}

type Tickers struct {
  SocketContext *common.SocketContext
  Hub           interface{}
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

var (
  globalNatsSub  *nats.Subscription
  globalNatsOnce sync.Once
)

func NewTickers(socketContext *common.SocketContext, hub interface{}) *Tickers {
  t := &Tickers{
    SocketContext: socketContext,
    Hub:           hub,
  }
  t.initGlobalNats()
  return t
}

func (h *Tickers) initGlobalNats() {
  if h.SocketContext == nil || h.SocketContext.Nats == nil {
    return
  }
  globalNatsOnce.Do(func() {
    sub, err := h.SocketContext.Nats.Subscribe(config.NATS_TICKERS_UPDATE, func(msg *nats.Msg) {
      var payload TickersUpdatePayload
      if err := json.Unmarshal(msg.Data, &payload); err != nil {
        return
      }

      topic := fmt.Sprintf("binance:spot:tickers:%s", payload.Symbol)
      pushBytes, _ := json.Marshal(map[string]interface{}{
        "event": "ticker",
        "topic": topic,
        "data":  payload,
      })

      if hubSub, ok := h.Hub.(HubSubscriber); ok {
        hubSub.Broadcast(topic, pushBytes)
      }
    })
    if err == nil {
      globalNatsSub = sub
    }
  })
}

func (h *Tickers) Subscribe(client interface{}, reqRaw interface{}) error {
  symbols, err := extractSymbols(reqRaw)
  if err != nil {
    return err
  }

  if hubSub, ok := h.Hub.(HubSubscriber); ok {
    for _, symbol := range symbols {
      topic := fmt.Sprintf("binance:spot:tickers:%s", symbol)
      hubSub.Subscribe(client, topic)

      if h.SocketContext != nil && h.SocketContext.Rdb != nil {
        redisKey := fmt.Sprintf(config.REDIS_KEY_TICKERS, symbol)
        data, err := h.SocketContext.Rdb.HGetAll(h.SocketContext.Ctx, redisKey).Result()
        if err == nil && len(data) > 0 {
          open, _ := strconv.ParseFloat(data["open"], 64)
          price, _ := strconv.ParseFloat(data["price"], 64)
          high, _ := strconv.ParseFloat(data["high"], 64)
          low, _ := strconv.ParseFloat(data["low"], 64)
          volume, _ := strconv.ParseFloat(data["volume"], 64)
          quota, _ := strconv.ParseFloat(data["quota"], 64)
          timestamp, _ := strconv.ParseInt(data["timestamp"], 10, 64)

          payload := TickersUpdatePayload{
            Symbol:    symbol,
            Open:      open,
            Price:     price,
            High:      high,
            Low:       low,
            Volume:    volume,
            Quota:     quota,
            Timestamp: timestamp,
          }

          pushBytes, _ := json.Marshal(map[string]interface{}{
            "event": "ticker",
            "topic": topic,
            "data":  payload,
          })

          if sender, ok := client.(ClientSender); ok {
            sender.SendBytes(pushBytes)
          }
        }
      }
    }
  }
  return nil
}

func (h *Tickers) Unsubscribe(client interface{}, reqRaw interface{}) error {
  symbols, err := extractSymbols(reqRaw)
  if err != nil {
    return err
  }

  if hubSub, ok := h.Hub.(HubSubscriber); ok {
    for _, symbol := range symbols {
      topic := fmt.Sprintf("binance:spot:tickers:%s", symbol)
      hubSub.Unsubscribe(client, topic)
    }
  }
  return nil
}

func extractSymbols(reqRaw interface{}) ([]string, error) {
  type RequestData struct {
    Symbols []string `json:"symbols"`
  }

  bytes, err := json.Marshal(reqRaw)
  if err != nil {
    return nil, errors.New("invalid request format")
  }

  var req RequestData
  if err := json.Unmarshal(bytes, &req); err != nil || len(req.Symbols) == 0 {
    return nil, errors.New("symbols is empty or invalid")
  }

  symbols := make([]string, len(req.Symbols))
  for i, s := range req.Symbols {
    symbols[i] = strings.ToUpper(strings.TrimSpace(s))
  }
  return symbols, nil
}
