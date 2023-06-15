package binance

import (
  "github.com/hibiken/asynq"
  "taoniu.local/cryptos/queue/asynq/workers/binance/futures"
)

type Futures struct{}

func NewFutures() *Futures {
  return &Futures{}
}

func (h *Futures) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("binance:futures:tickers:flush", futures.NewTickers().Flush)
  mux.HandleFunc("binance:futures:klines:flush", futures.NewKlines().Flush)

  return nil
}
