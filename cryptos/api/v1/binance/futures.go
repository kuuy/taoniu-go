package binance

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/futures"
)

func NewFuturesRouter() http.Handler {
  r := chi.NewRouter()
  r.Mount("/tickers", futures.NewTickersRouter())
  r.Mount("/indicators", futures.NewIndicatorsRouter())
  r.Mount("/strategies", futures.NewStrategiesRouter())
  r.Mount("/plans", futures.NewPlansRouter())
  r.Mount("/orders", futures.NewOrdersRouter())
  r.Mount("/positions", futures.NewPositionsRouter())
  r.Mount("/scalping", futures.NewScalpingRouter())
  r.Mount("/triggers", futures.NewTriggersRouter())
  r.Mount("/tradings", futures.NewTradingsouter())
  r.Mount("/tradingview", futures.NewTradingViewRouter())
  return r
}
