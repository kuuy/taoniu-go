package tradingview

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/dydx/tradingview/templates"
)

func NewTemplatesRouter() http.Handler {
  r := chi.NewRouter()
  r.Mount("/study", templates.NewStudyRouter())
  r.Mount("/drawing", templates.NewDrawingRouter())
  return r
}
