package margin

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/margin/cross"
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
  cross.NewAccount(h.AnsqContext).Register()
  cross.NewOrders(h.AnsqContext).Register()
  cross.NewPositions(h.AnsqContext).Register()
  cross.NewTradings(h.AnsqContext).Register()
  return nil
}
