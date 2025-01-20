package spot

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot/tradings"
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
  tradings.NewGambling(h.AnsqContext).Register()
  return nil
}
