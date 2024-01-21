package isolated

import (
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot/margin/isolated/tradings"
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
  h.AnsqContext.Mux.HandleFunc("binance:spot:margin:isolated:tradings:fishers:flush", tradings.NewFishers().Flush)
  h.AnsqContext.Mux.HandleFunc("binance:spot:margin:isolated:tradings:fishers:place", tradings.NewFishers().Place)
  return nil
}
