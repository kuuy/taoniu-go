package margin

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/spot/analysis/margin/isolated"
  "taoniu.local/cryptos/common"
)

func NewIsolatedRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/tradings", isolated.NewTradingsRouter(apiContext))
  return r
}
