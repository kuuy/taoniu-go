package workers

import (
	"context"
	"github.com/hibiken/asynq"
	"taoniu.local/cryptos/queue/workers/tradingview"
)

type Tradingview struct{}

func NewTradingview() *Tradingview {
	return &Tradingview{}
}

func (h *Tradingview) Register(mux *asynq.ServeMux) error {
	mux.HandleFunc("tradingview:analysis:flush", func(ctx context.Context, t *asynq.Task) error {
		return tradingview.NewAnalysis().Flush(ctx, t)
	})
	return nil
}
