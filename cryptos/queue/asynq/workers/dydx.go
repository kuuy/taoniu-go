package workers

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/dydx"
)

type Dydx struct {
  AnsqContext *common.AnsqServerContext
}

func NewDydx(ansqContext *common.AnsqServerContext) *Dydx {
  return &Dydx{
    AnsqContext: ansqContext,
  }
}

func (h *Dydx) Register() error {
  dydx.NewTickers(h.AnsqContext).Register()
  dydx.NewOrderbook(h.AnsqContext).Register()
  dydx.NewKlines(h.AnsqContext).Register()
  dydx.NewIndicators(h.AnsqContext).Register()
  dydx.NewStrategies(h.AnsqContext).Register()
  dydx.NewPlans(h.AnsqContext).Register()
  dydx.NewAccount(h.AnsqContext).Register()
  dydx.NewOrders(h.AnsqContext).Register()
  dydx.NewTradings(h.AnsqContext).Register()
  return nil
}
