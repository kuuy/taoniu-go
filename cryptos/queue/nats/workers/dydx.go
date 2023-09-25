package workers

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/nats/workers/dydx"
)

type Dydx struct {
  NatsContext *common.NatsContext
}

func NewDydx(natsContext *common.NatsContext) *Dydx {
  return &Dydx{
    NatsContext: natsContext,
  }
}

func (h *Dydx) Subscribe() error {
  dydx.NewPatterns(h.NatsContext).Subscribe()
  dydx.NewIndicators(h.NatsContext).Subscribe()
  dydx.NewStrategies(h.NatsContext).Subscribe()
  dydx.NewPlans(h.NatsContext).Subscribe()
  dydx.NewTrades(h.NatsContext).Subscribe()
  dydx.NewOrders(h.NatsContext).Subscribe()
  dydx.NewScalping(h.NatsContext).Subscribe()
  dydx.NewTradings(h.NatsContext).Subscribe()
  return nil
}
