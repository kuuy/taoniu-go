package spot

import (
  "encoding/json"
  "fmt"
  "log"

  "github.com/eclipse/paho.golang/paho"
  "github.com/nats-io/nats.go"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
)

type Orders struct {
  MqttContext *common.MqttContext
}

func NewOrders(mqttContext *common.MqttContext) *Orders {
  h := &Orders{
    MqttContext: mqttContext,
  }
  return h
}

func (h *Orders) Subscribe() error {
  h.MqttContext.Nats.Subscribe(config.NATS_ORDERS_UPDATE, h.Update)
  return nil
}

func (h *Orders) Update(m *nats.Msg) {
  var payload *OrdersUpdatePayload
  json.Unmarshal(m.Data, &payload)

  props := &paho.PublishProperties{}

  if _, err := h.MqttContext.Conn.Publish(h.MqttContext.Ctx, &paho.Publish{
    Topic:      fmt.Sprintf(config.MQTT_TOPICS_ORDERS, payload.Symbol),
    QoS:        0,
    Payload:    m.Data,
    Properties: props,
  }); err != nil {
    log.Fatalf("failed to publish message: %s", err)
  }
}
