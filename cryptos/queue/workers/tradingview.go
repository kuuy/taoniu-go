package workers

import (
	"github.com/hibiken/asynq"
	"taoniu.local/cryptos/queue/workers/tradingview"
)

type Tradingview struct{}

func NewTradingview() *Tradingview {
	return &Tradingview{}
}

func (h *Tradingview) Register(mux *asynq.ServeMux) error {
	mux.HandleFunc("tradingview:analysis:flush", tradingview.NewAnalysis().Flush)
	return nil
}
