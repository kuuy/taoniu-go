package binance

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api/v1/binance/margin"
  "taoniu.local/cryptos/common"
)

func NewMarginRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/cross", margin.NewCrossRouter(apiContext))
  r.Mount("/isolated", margin.NewIsolatedRouter(apiContext))
  return r
}
