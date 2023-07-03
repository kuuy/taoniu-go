package futures

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance/futures/strategies"
)

type Strategies struct{}

func NewStrategies() *Strategies {
  return &Strategies{}
}

func (h *Strategies) Register(mux *asynq.ServeMux) error {
  strategies.NewDaily().Register(mux)
  strategies.NewMinutely().Register(mux)
  return nil
}
