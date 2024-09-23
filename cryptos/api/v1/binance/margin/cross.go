package margin

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api/v1/binance/margin/cross"
  "taoniu.local/cryptos/common"
)

func NewCrossRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/analysis", cross.NewAnalysisRouter(apiContext))
  r.Mount("/orders", cross.NewOrdersRouter(apiContext))
  return r
}
