package binance

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/nats/workers/binance/futures"
)

type Futures struct {
  NatsContext *common.NatsContext
}

func NewFutures(natsContext *common.NatsContext) *Futures {
  return &Futures{
    NatsContext: natsContext,
  }
}

func (h *Futures) Subscribe() error {
  futures.NewTickers(h.NatsContext).Subscribe()
  //futures.NewKlines(h.NatsContext).Subscribe()
  //futures.NewPatterns(h.NatsContext).Subscribe()
  futures.NewIndicators(h.NatsContext).Subscribe()
  futures.NewStrategies(h.NatsContext).Subscribe()
  futures.NewPlans(h.NatsContext).Subscribe()
  futures.NewAccount(h.NatsContext).Subscribe()
  futures.NewOrders(h.NatsContext).Subscribe()
  futures.NewScalping(h.NatsContext).Subscribe()
  futures.NewTradings(h.NatsContext).Subscribe()
  return nil
}
