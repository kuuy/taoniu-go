package binance

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance/futures"
)

type Futures struct{}

func NewFutures() *Futures {
  return &Futures{}
}

func (h *Futures) Register(mux *asynq.ServeMux) error {
  futures.NewTickers().Register(mux)
  futures.NewKlines().Register(mux)
  futures.NewDepth().Register(mux)
  futures.NewIndicators().Register(mux)
  futures.NewStrategies().Register(mux)
  futures.NewPlans().Register(mux)
  futures.NewAccount().Register(mux)
  futures.NewOrders().Register(mux)
  futures.NewTradings().Register(mux)
  return nil
}
