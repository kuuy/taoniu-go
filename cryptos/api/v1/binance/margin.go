package binance

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/margin"
  "taoniu.local/cryptos/common"
)

func NewMarginRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/cross", margin.NewCrossRouter(apiContext))
  r.Mount("/cross", margin.NewIsolatedRouter(apiContext))
  return r
}
