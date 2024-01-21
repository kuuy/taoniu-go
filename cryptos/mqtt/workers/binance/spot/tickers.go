package spot

import (
  "encoding/json"
  "fmt"
  "github.com/eclipse/paho.golang/paho"
  "log"

  "github.com/nats-io/nats.go"
  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
)

type Tickers struct {
  MqttContext *common.MqttContext
}

func NewTickers(mqttContext *common.MqttContext) *Tickers {
  h := &Tickers{
    MqttContext: mqttContext,
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
  h.MqttContext.Nats.Subscribe(config.NATS_TICKERS_UPDATE, h.Update)
  return nil
}

func (h *Tickers) Update(m *nats.Msg) {
  var payload *TickersUpdatePayload
  json.Unmarshal(m.Data, &payload)

  props := &paho.PublishProperties{}

  if _, err := h.MqttContext.Conn.Publish(h.MqttContext.Ctx, &paho.Publish{
    Topic:      fmt.Sprintf(config.MQTT_TOPICS_TICKERS, payload.Symbol),
    QoS:        0,
    Payload:    m.Data,
    Properties: props,
  }); err != nil {
    log.Fatalf("failed to publish message: %s", err)
  }
}
