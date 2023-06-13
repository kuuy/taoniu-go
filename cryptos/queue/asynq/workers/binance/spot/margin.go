package spot

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot/margin"
)

type Margin struct{}

func NewMargin() *Margin {
  return &Margin{}
}

func (h *Margin) Register(mux *asynq.ServeMux) error {
  margin.NewCross().Register(mux)
  margin.NewIsolated().Register(mux)
  return nil
}
