package analysis

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/spot/analysis/tradings"
  "taoniu.local/cryptos/common"
)

func NewTradingsRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/scalping", tradings.NewScalpingRouter(apiContext))
  r.Mount("/triggers", tradings.NewTriggersRouter(apiContext))
  return r
}
