package binance

import (
	"github.com/hibiken/asynq"
	"taoniu.local/cryptos/queue/workers/binance/spot"
)

type Spot struct{}

func NewSpot() *Spot {
	return &Spot{}
}

func (h *Spot) Register(mux *asynq.ServeMux) error {
	tickers := spot.NewTickers()
	mux.HandleFunc("binance:spot:tickers:flush", tickers.Flush)
	return nil
}
