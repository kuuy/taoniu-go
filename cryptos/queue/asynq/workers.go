package asynq

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers"
)

type Workers struct{}

func NewWorkers() *Workers {
  return &Workers{}
}

func (h *Workers) Register(mux *asynq.ServeMux) error {
  workers.NewBinance().Register(mux)
  workers.NewDydx().Register(mux)
  workers.NewTradingview().Register(mux)
  return nil
}
