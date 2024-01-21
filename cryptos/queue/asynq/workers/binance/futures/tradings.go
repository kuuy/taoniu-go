package futures

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/futures/tradings"
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
  h.AnsqContext.Mux.HandleFunc("binance:futures:tradings:scalping:place", tradings.NewScalping().Place)
  h.AnsqContext.Mux.HandleFunc("binance:futures:tradings:scalping:flush", tradings.NewScalping().Flush)
  h.AnsqContext.Mux.HandleFunc("binance:futures:tradings:triggers:place", tradings.NewTriggers().Place)
  h.AnsqContext.Mux.HandleFunc("binance:futures:tradings:triggers:flush", tradings.NewTriggers().Flush)
  return nil
}
