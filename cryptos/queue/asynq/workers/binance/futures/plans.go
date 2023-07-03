package futures

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance/futures/plans"
)

type Plans struct{}

func NewPlans() *Plans {
  return &Plans{}
}

func (h *Plans) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:futures:plans:1d:flush", plans.NewDaily().Flush)
  mux.HandleFunc("binance:futures:plans:1m:flush", plans.NewMinutely().Flush)
  return nil
}
