package v1

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/api/v1/binance"
  "taoniu.local/cryptos/common"
)

func NewBinanceRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Use(api.Authenticator)
  r.Mount("/spot", binance.NewSpotRouter(apiContext))
  r.Mount("/futures", binance.NewFuturesRouter(apiContext))
  return r
}
