package futures

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/futures/tradings"
  "taoniu.local/cryptos/common"
)

func NewTradingsRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/scalping", tradings.NewScalpingRouter(apiContext))
  return r
}
