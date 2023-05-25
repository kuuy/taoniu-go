package margin

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot/margin/isolated"
)

type Isolated struct{}

func NewIsolated() *Isolated {
  return &Isolated{}
}

func (h *Isolated) Register(mux *asynq.ServeMux) error {
  isolated.NewTradings().Register(mux)
  return nil
}
