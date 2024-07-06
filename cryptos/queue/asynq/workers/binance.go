package workers

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance"
)

type Binance struct {
  AnsqContext *common.AnsqServerContext
}

func NewBinance(ansqContext *common.AnsqServerContext) *Binance {
  return &Binance{
    AnsqContext: ansqContext,
  }
}

func (h *Binance) Register() error {
  binance.NewSpot(h.AnsqContext).Register()
  binance.NewMargin(h.AnsqContext).Register()
  binance.NewFutures(h.AnsqContext).Register()
  return nil
}
