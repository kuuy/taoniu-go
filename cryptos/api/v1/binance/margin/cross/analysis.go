package cross

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api/v1/binance/margin/cross/analysis"
  "taoniu.local/cryptos/common"
)

func NewAnalysisRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/tradings", analysis.NewTradingsRouter(apiContext))
  return r
}
