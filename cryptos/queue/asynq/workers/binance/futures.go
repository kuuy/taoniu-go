package binance

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/futures"
)

type Futures struct {
  AnsqContext *common.AnsqServerContext
}

func NewFutures(ansqContext *common.AnsqServerContext) *Futures {
  return &Futures{
    AnsqContext: ansqContext,
  }
}

func (h *Futures) Register() error {
  futures.NewTickers(h.AnsqContext).Register()
  futures.NewKlines(h.AnsqContext).Register()
  futures.NewDepth(h.AnsqContext).Register()
  futures.NewIndicators(h.AnsqContext).Register()
  futures.NewStrategies(h.AnsqContext).Register()
  futures.NewPlans(h.AnsqContext).Register()
  futures.NewAccount(h.AnsqContext).Register()
  futures.NewOrders(h.AnsqContext).Register()
  futures.NewTradings(h.AnsqContext).Register()
  return nil
}
