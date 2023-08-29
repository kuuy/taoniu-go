package v1

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/dydx"
)

func NewDydxRouter() http.Handler {
  r := chi.NewRouter()
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
