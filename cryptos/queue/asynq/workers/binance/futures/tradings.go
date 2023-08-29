package futures

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance/futures/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
  return &Tradings{}
}

func (h *Tradings) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:futures:tradings:scalping:place", tradings.NewScalping().Place)
  mux.HandleFunc("binance:futures:tradings:scalping:flush", tradings.NewScalping().Flush)
  mux.HandleFunc("binance:futures:tradings:triggers:place", tradings.NewTriggers().Place)
  mux.HandleFunc("binance:futures:tradings:triggers:flush", tradings.NewTriggers().Flush)
  return nil
}
