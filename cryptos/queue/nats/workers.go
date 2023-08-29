package nats

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/nats/workers"
)

type Workers struct {
  NatsContext *common.NatsContext
}

func NewWorkers(natsContext *common.NatsContext) *Workers {
  return &Workers{
    NatsContext: natsContext,
  }
}

func (h *Workers) Subscribe() error {
  workers.NewBinance(h.NatsContext).Subscribe()
  workers.NewDydx(h.NatsContext).Subscribe()
  return nil
}
