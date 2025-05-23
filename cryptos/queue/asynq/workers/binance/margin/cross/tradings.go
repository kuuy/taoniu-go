package cross

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/margin/cross/tradings"
)

type Tradings struct {
  AnsqContext *common.AnsqServerContext
}

func NewTradings(ansqContext *common.AnsqServerContext) *Tradings {
  return &Tradings{
    AnsqContext: ansqContext,
  }
}

func (h *Tradings) Register() error {
  tradings.NewScalping(h.AnsqContext).Register()
  return nil
}
