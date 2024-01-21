package asynq

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers"
)

type Workers struct {
  AnsqContext *common.AnsqServerContext
}

func NewWorkers(ansqContext *common.AnsqServerContext) *Workers {
  return &Workers{
    AnsqContext: ansqContext,
  }
}

func (h *Workers) Register() error {
  workers.NewBinance(h.AnsqContext).Register()
  workers.NewDydx(h.AnsqContext).Register()
  workers.NewTradingview(h.AnsqContext).Register()
  return nil
}
