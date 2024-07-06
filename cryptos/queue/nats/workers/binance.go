package workers

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/nats/workers/binance"
)

type Binance struct {
  NatsContext *common.NatsContext
}

func NewBinance(natsContext *common.NatsContext) *Binance {
  return &Binance{
    NatsContext: natsContext,
  }
}

func (h *Binance) Subscribe() error {
  binance.NewSpot(h.NatsContext).Subscribe()
  binance.NewMargin(h.NatsContext).Subscribe()
  binance.NewFutures(h.NatsContext).Subscribe()
  return nil
}
