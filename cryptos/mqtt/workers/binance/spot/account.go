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

type Account struct {
  MqttContext *common.MqttContext
}

func NewAccount(mqttContext *common.MqttContext) *Account {
  h := &Account{
    MqttContext: mqttContext,
  }
  return h
}

func (h *Account) Subscribe() error {
  h.MqttContext.Nats.Subscribe(config.NATS_ACCOUNT_UPDATE, h.Update)
  return nil
}

func (h *Account) Update(m *nats.Msg) {
  var payload *AccountUpdatePayload
  json.Unmarshal(m.Data, &payload)

  props := &paho.PublishProperties{}

  if _, err := h.MqttContext.Conn.Publish(h.MqttContext.Ctx, &paho.Publish{
    Topic:      fmt.Sprintf(config.MQTT_TOPICS_ACCOUNT, payload.Asset),
    QoS:        0,
    Payload:    m.Data,
    Properties: props,
  }); err != nil {
    log.Fatalf("failed to publish message: %s", err)
  }
}
