package binance

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/mqtt/workers/binance/futures"
)

type Futures struct {
  MqttContext *common.MqttContext
}

func NewFutures(mqttContext *common.MqttContext) *Futures {
  return &Futures{
    MqttContext: mqttContext,
  }
}

func (h *Futures) Subscribe() error {
  futures.NewAccount(h.MqttContext).Subscribe()
  futures.NewOrders(h.MqttContext).Subscribe()
  futures.NewTickers(h.MqttContext).Subscribe()
  return nil
}
