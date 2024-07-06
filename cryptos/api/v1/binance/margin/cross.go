package margin

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/margin/cross"
  "taoniu.local/cryptos/common"
)

func NewCrossRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/orders", cross.NewOrdersRouter(apiContext))
  return r
}
