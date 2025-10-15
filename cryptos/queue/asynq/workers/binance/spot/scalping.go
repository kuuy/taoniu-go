package spot

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot/scalping"
)

type Scalping struct {
  AnsqContext *common.AnsqServerContext
}

func NewScalping(ansqContext *common.AnsqServerContext) *Scalping {
  return &Scalping{
    AnsqContext: ansqContext,
  }
}

func (h *Scalping) Register() error {
  scalping.NewPlans(h.AnsqContext).Register()
  return nil
}
