package spot

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot/margin"
)

type Margin struct {
  AnsqContext *common.AnsqServerContext
}

func NewMargin(ansqContext *common.AnsqServerContext) *Margin {
  return &Margin{
    AnsqContext: ansqContext,
  }
}

func (h *Margin) Register() error {
  margin.NewCross(h.AnsqContext).Register()
  margin.NewIsolated(h.AnsqContext).Register()
  return nil
}
