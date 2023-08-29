package dydx

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/dydx/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
  return &Tradings{}
}

func (h *Tradings) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("dydx:tradings:scalping:place", tradings.NewScalping().Place)
  mux.HandleFunc("dydx:tradings:scalping:flush", tradings.NewScalping().Flush)
  mux.HandleFunc("dydx:tradings:triggers:place", tradings.NewTriggers().Place)
  mux.HandleFunc("dydx:tradings:triggers:flush", tradings.NewTriggers().Flush)
  return nil
}
