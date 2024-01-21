package tradings

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/spot/analysis/tradings/fishers"
  "taoniu.local/cryptos/common"
)

func NewFishersRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/grids", fishers.NewGridsRouter(apiContext))
  return r
}
