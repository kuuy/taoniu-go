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
  h.AnsqContext.Mux.HandleFunc("binance:spot:tradings:launchpad:place", tradings.NewLaunchpad().Place)
  h.AnsqContext.Mux.HandleFunc("binance:spot:tradings:launchpad:flush", tradings.NewLaunchpad().Flush)
  h.AnsqContext.Mux.HandleFunc("binance:spot:tradings:scalping:place", tradings.NewScalping().Place)
  h.AnsqContext.Mux.HandleFunc("binance:spot:tradings:scalping:flush", tradings.NewScalping().Flush)
  h.AnsqContext.Mux.HandleFunc("binance:spot:tradings:triggers:place", tradings.NewTriggers().Place)
  h.AnsqContext.Mux.HandleFunc("binance:spot:tradings:triggers:flush", tradings.NewTriggers().Flush)
  return nil
}
