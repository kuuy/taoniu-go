package spot

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/nats/workers/binance/spot/tradings"
)

type Tradings struct {
  NatsContext *common.NatsContext
}

func NewTradings(natsContext *common.NatsContext) *Tradings {
  return &Tradings{
    NatsContext: natsContext,
  }
}

func (h *Tradings) Subscribe() error {
  tradings.NewScalping(h.NatsContext).Subscribe()
  return nil
}
