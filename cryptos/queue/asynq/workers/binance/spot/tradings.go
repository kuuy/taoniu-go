package spot

import (
  "github.com/hibiken/asynq"
  tradings2 "taoniu.local/cryptos/queue/asynq/workers/binance/spot/tradings"
)

type Tradings struct{}

func NewTradings() *Tradings {
  return &Tradings{}
}

func (h *Tradings) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:spot:tradings:fishers:flush", tradings2.NewFishers().Flush)
  mux.HandleFunc("binance:spot:tradings:fishers:place", tradings2.NewFishers().Place)
  mux.HandleFunc("binance:spot:tradings:scalping:flush", tradings2.NewScalping().Flush)
  return nil
}
