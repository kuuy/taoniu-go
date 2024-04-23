package futures

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/futures/analysis"
  "taoniu.local/cryptos/common"
)

func NewAnalysisRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/tradings", analysis.NewTradingsRouter(apiContext))
  return r
}
