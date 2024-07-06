package binance

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/nats/workers/binance/margin"
)

type Margin struct {
  NatsContext *common.NatsContext
}

func NewMargin(natsContext *common.NatsContext) *Margin {
  return &Margin{
    NatsContext: natsContext,
  }
}

func (h *Margin) Subscribe() error {
  margin.NewCross(h.NatsContext).Subscribe()
  return nil
}
