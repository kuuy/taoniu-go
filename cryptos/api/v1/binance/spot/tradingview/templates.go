package tradingview

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/spot/tradingview/templates"
  "taoniu.local/cryptos/common"
)

func NewTemplatesRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/study", templates.NewStudyRouter(apiContext))
  r.Mount("/drawing", templates.NewDrawingRouter(apiContext))
  return r
}
