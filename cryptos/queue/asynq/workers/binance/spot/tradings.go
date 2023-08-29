package spot

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
  return &Tradings{}
}

func (h *Tradings) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:spot:tradings:launchpad:place", tradings.NewLaunchpad().Place)
  mux.HandleFunc("binance:spot:tradings:launchpad:flush", tradings.NewLaunchpad().Flush)
  mux.HandleFunc("binance:spot:tradings:scalping:place", tradings.NewScalping().Place)
  mux.HandleFunc("binance:spot:tradings:scalping:flush", tradings.NewScalping().Flush)
  mux.HandleFunc("binance:spot:tradings:triggers:place", tradings.NewTriggers().Place)
  mux.HandleFunc("binance:spot:tradings:triggers:flush", tradings.NewTriggers().Flush)
  return nil
}
