package workers

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance"
)

type Binance struct{}

func NewBinance() *Binance {
  return &Binance{}
}

func (h *Binance) Register(mux *asynq.ServeMux) error {
  binance.NewSpot().Register(mux)
  return nil
}
