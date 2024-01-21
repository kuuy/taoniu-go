package dydx

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/dydx/tradingview"
  "taoniu.local/cryptos/common"
)

func NewTradingViewRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/datafeed", tradingview.NewDatafeedRouter(apiContext))
  r.Mount("/charts", tradingview.NewChartsRouter(apiContext))
  r.Mount("/templates", tradingview.NewTemplatesRouter(apiContext))
  return r
}
