package spot

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api/v1/binance/spot/margin"
  "taoniu.local/cryptos/common"
)

func NewMarginRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/orders", margin.NewOrdersRouter(apiContext))
  r.Mount("/isolated", margin.NewIsolatedRouter(apiContext))
  return r
}
