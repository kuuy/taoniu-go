package spot

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api/v1/binance/spot/tradings"
  "taoniu.local/cryptos/common"
)

func NewTradingsRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/scalping", tradings.NewScalpingRouter(apiContext))
  return r
}
