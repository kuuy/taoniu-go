package binance

import (
	"context"
	"github.com/hibiken/asynq"
	"taoniu.local/cryptos/queue/workers/binance/spot"
)

type Spot struct{}

func NewSpot() *Spot {
	return &Spot{}
}

func (h *Spot) Register(mux *asynq.ServeMux) error {
	mux.HandleFunc("binance:spot:depth:flush", func(ctx context.Context, t *asynq.Task) error {
		return spot.NewDepth().Flush(ctx, t)
	})
	mux.HandleFunc("binance:spot:tickers:flush", func(ctx context.Context, t *asynq.Task) error {
		return spot.NewTickers().Flush(ctx, t)
	})
	mux.HandleFunc("binance:spot:klines:flush", func(ctx context.Context, t *asynq.Task) error {
		return spot.NewKlines().Flush(ctx, t)
	})
	return nil
}
