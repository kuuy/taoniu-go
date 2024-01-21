package margin

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot/margin/cross"
)

type Cross struct {
  AnsqContext *common.AnsqServerContext
}

func NewCross(ansqContext *common.AnsqServerContext) *Cross {
  return &Cross{
    AnsqContext: ansqContext,
  }
}

func (h *Cross) Register() error {
  cross.NewTradings(h.AnsqContext).Register()
  return nil
}
