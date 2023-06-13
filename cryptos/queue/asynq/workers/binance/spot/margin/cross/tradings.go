package cross

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot/margin/cross/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
  return &Tradings{}
}

func (h *Tradings) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:spot:margin:cross:tradings:triggers:flush", tradings.NewTriggers().Flush)
  mux.HandleFunc("binance:spot:margin:cross:tradings:triggers:place", tradings.NewTriggers().Place)
  return nil
}
