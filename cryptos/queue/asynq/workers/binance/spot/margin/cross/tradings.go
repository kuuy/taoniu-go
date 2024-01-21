package cross

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot/margin/cross/tradings"
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
  h.AnsqContext.Mux.HandleFunc("binance:spot:margin:cross:tradings:triggers:flush", tradings.NewTriggers().Flush)
  h.AnsqContext.Mux.HandleFunc("binance:spot:margin:cross:tradings:triggers:place", tradings.NewTriggers().Place)
  return nil
}
