package spot

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/spot/tradingview"
)

func NewTradingViewRouter() http.Handler {
  r := chi.NewRouter()
  r.Mount("/datafeed", tradingview.NewDatafeedRouter())
  return r
}
