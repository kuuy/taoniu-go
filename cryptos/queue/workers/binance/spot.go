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
	mux.HandleFunc("binance:spot:depth:flush", spot.NewDepth().Flush)
	mux.HandleFunc("binance:spot:tickers:flush", spot.NewTickers().Flush)
	return nil
}
