package binance

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/mqtt/workers/binance/spot"
)

type Spot struct {
  MqttContext *common.MqttContext
}

func NewSpot(mqttContext *common.MqttContext) *Spot {
  return &Spot{
    MqttContext: mqttContext,
  }
}

func (h *Spot) Subscribe() error {
  spot.NewTickers(h.MqttContext).Subscribe()
  return nil
}
