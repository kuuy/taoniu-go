package spot

import (
	"github.com/hibiken/asynq"
	"taoniu.local/cryptos/queue/workers/binance/spot/margin"
)

type Margin struct{}

func NewMargin() *Margin {
	return &Margin{}
}

func (h *Margin) Register(mux *asynq.ServeMux) error {
	margin.NewIsolated().Register(mux)
	return nil
}
