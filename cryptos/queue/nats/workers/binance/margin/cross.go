package margin

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/nats/workers/binance/margin/cross"
)

type Cross struct {
  NatsContext *common.NatsContext
}

func NewCross(natsContext *common.NatsContext) *Cross {
  return &Cross{
    NatsContext: natsContext,
  }
}

func (h *Cross) Subscribe() error {
  cross.NewAccount(h.NatsContext).Subscribe()
  cross.NewScalping(h.NatsContext).Subscribe()
  return nil
}
