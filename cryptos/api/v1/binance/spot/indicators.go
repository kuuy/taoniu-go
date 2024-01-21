package spot

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/spot/indicators"
  "taoniu.local/cryptos/common"
)

func NewIndicatorsRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/daily", indicators.NewDailyRouter(apiContext))
  return r
}
