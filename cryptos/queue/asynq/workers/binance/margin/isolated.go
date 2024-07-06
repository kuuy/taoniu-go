package margin

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/margin/isolated"
)

type Isolated struct {
  AnsqContext *common.AnsqServerContext
}

func NewIsolated(ansqContext *common.AnsqServerContext) *Isolated {
  return &Isolated{
    AnsqContext: ansqContext,
  }
}

func (h *Isolated) Register() error {
  isolated.NewTradings(h.AnsqContext).Register()
  return nil
}
