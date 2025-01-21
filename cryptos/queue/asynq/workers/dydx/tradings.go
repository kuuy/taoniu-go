package dydx

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/dydx/tradings"
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
  h.AnsqContext.Mux.HandleFunc("dydx:tradings:scalping:place", tradings.NewScalping(h.AnsqContext).Place)
  h.AnsqContext.Mux.HandleFunc("dydx:tradings:scalping:flush", tradings.NewScalping(h.AnsqContext).Flush)
  return nil
}
