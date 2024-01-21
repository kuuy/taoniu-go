package v1

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/api/v1/tradingview"
  "taoniu.local/cryptos/common"
)

func NewTradingviewRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Use(api.Authenticator)
  r.Mount("/analysis", tradingview.NewAnalysisRouter(apiContext))
  r.Mount("/summary/{exchange:[A-Z0-9]{1,20}},{symbol:[A-Z0-9]{1,20}},{interval:[a-zA-Z0-9]{2}}", tradingview.NewSummaryRouter(apiContext))

  return r
}
