package spot

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/spot/plans"
  "taoniu.local/cryptos/common"
)

func NewPlansRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/daily", plans.NewDailyRouter(apiContext))
  return r
}
