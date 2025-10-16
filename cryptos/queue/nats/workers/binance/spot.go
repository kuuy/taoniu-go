package binance

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/nats/workers/binance/spot"
)

type Spot struct {
  NatsContext *common.NatsContext
}

func NewSpot(natsContext *common.NatsContext) *Spot {
  return &Spot{
    NatsContext: natsContext,
  }
}

func (h *Spot) Subscribe() error {
  spot.NewTickers(h.NatsContext).Subscribe()
  //spot.NewKlines(h.NatsContext).Subscribe()
  spot.NewIndicators(h.NatsContext).Subscribe()
  spot.NewStrategies(h.NatsContext).Subscribe()
  spot.NewPlans(h.NatsContext).Subscribe()
  spot.NewAccount(h.NatsContext).Subscribe()
  spot.NewOrders(h.NatsContext).Subscribe()
  spot.NewScalping(h.NatsContext).Subscribe()
  spot.NewTradings(h.NatsContext).Subscribe()
  return nil
}
