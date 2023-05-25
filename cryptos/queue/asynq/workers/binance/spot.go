package binance

import (
  "github.com/hibiken/asynq"
  spot2 "taoniu.local/cryptos/queue/asynq/workers/binance/spot"
)

type Spot struct{}

func NewSpot() *Spot {
  return &Spot{}
}

func (h *Spot) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:spot:depth:flush", spot2.NewDepth().Flush)
  mux.HandleFunc("binance:spot:tickers:flush", spot2.NewTickers().Flush)
  mux.HandleFunc("binance:spot:klines:flush", spot2.NewKlines().Flush)

  spot2.NewTradings().Register(mux)
  spot2.NewMargin().Register(mux)

  return nil
}
