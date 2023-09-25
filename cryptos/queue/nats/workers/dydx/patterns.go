package dydx

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/nats/workers/dydx/patterns"
)

type Patterns struct {
  NatsContext *common.NatsContext
}

func NewPatterns(natsContext *common.NatsContext) *Patterns {
  return &Patterns{
    NatsContext: natsContext,
  }
}

func (h *Patterns) Subscribe() error {
  patterns.NewCandlesticks(h.NatsContext).Subscribe()
  return nil
}
