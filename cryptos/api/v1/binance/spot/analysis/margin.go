package analysis

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/spot/analysis/margin"
  "taoniu.local/cryptos/common"
)

func NewMarginRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/isolated", margin.NewIsolatedRouter(apiContext))
  return r
}
