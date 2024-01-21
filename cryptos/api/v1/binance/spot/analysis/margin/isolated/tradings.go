package isolated

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/spot/analysis/margin/isolated/tradings"
  "taoniu.local/cryptos/common"
)

func NewTradingsRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/fishers", tradings.NewFishersRouter(apiContext))
  return r
}
