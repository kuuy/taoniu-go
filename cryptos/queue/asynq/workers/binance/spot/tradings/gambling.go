package tradings

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot/tradings/gambling"
)

type Gambling struct {
  AnsqContext *common.AnsqServerContext
}

func NewGambling(ansqContext *common.AnsqServerContext) *Gambling {
  return &Gambling{
    AnsqContext: ansqContext,
  }
}

func (h *Gambling) Register() error {
  gambling.NewAnt(h.AnsqContext).Register()
  gambling.NewScalping(h.AnsqContext).Register()
  return nil
}
