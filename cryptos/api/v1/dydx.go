package v1

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/api/v1/dydx"
)

func NewDydxRouter() http.Handler {
  r := chi.NewRouter()
  r.Use(api.Authenticator)
  r.Mount("/tickers", dydx.NewTickersRouter())
  r.Mount("/indicators", dydx.NewIndicatorsRouter())
  r.Mount("/plans", dydx.NewPlansRouter())
  r.Mount("/orders", dydx.NewOrdersRouter())
  r.Mount("/positions", dydx.NewPositionsRouter())
  r.Mount("/scalping", dydx.NewScalpingRouter())
  r.Mount("/triggers", dydx.NewTriggersRouter())
  r.Mount("/tradings", dydx.NewTradingsouter())
  r.Mount("/tradingview", dydx.NewTradingViewRouter())
  return r
}
