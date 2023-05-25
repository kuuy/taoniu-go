package asynq

import (
  "github.com/hibiken/asynq"
  workers2 "taoniu.local/cryptos/queue/asynq/workers"
)

type Workers struct{}

func NewWorkers() *Workers {
  return &Workers{}
}

func (h *Workers) Register(mux *asynq.ServeMux) error {
  workers2.NewBinance().Register(mux)
  workers2.NewTradingview().Register(mux)
  return nil
}
