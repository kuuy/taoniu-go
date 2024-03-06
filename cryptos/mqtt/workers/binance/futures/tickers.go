package futures

import (
  "encoding/json"
  "fmt"
  "log"

  "github.com/eclipse/paho.golang/paho"
  "github.com/nats-io/nats.go"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
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
