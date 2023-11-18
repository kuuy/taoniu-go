package workers

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/dydx"
)

type Dydx struct{}

func NewDydx() *Dydx {
  return &Dydx{}
}

func (h *Dydx) Register(mux *asynq.ServeMux) error {
  dydx.NewTickers().Register(mux)
  dydx.NewOrderbook().Register(mux)
  dydx.NewKlines().Register(mux)
  dydx.NewIndicators().Register(mux)
  dydx.NewStrategies().Register(mux)
  dydx.NewPlans().Register(mux)
  dydx.NewAccount().Register(mux)
  dydx.NewOrders().Register(mux)
  dydx.NewTradings().Register(mux)
  return nil
}
