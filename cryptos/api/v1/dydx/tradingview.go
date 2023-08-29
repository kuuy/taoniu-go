package dydx

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/dydx/tradingview"
)

func NewTradingViewRouter() http.Handler {
  r := chi.NewRouter()
  r.Mount("/datafeed", tradingview.NewDatafeedRouter())
  r.Mount("/charts", tradingview.NewChartsRouter())
  r.Mount("/templates", tradingview.NewTemplatesRouter())
  return r
}
