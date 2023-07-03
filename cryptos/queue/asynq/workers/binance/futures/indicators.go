package futures

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance/futures/indicators"
)

type Indicators struct{}

func NewIndicators() *Indicators {
  return &Indicators{}
}

func (h *Indicators) Register(mux *asynq.ServeMux) error {
  indicators.NewDaily().Register(mux)
  indicators.NewMinutely().Register(mux)
  return nil
}
