package margin

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance/spot/margin/cross"
)

type Cross struct{}

func NewCross() *Cross {
  return &Cross{}
}

func (h *Cross) Register(mux *asynq.ServeMux) error {
  cross.NewTradings().Register(mux)
  return nil
}
