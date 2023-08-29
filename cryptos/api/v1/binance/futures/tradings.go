package futures

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/futures/tradings"
)

func NewTradingsouter() http.Handler {
  r := chi.NewRouter()
  r.Mount("/scalping", tradings.NewScalpingRouter())
  r.Mount("/triggers", tradings.NewTriggersRouter())
  return r
}
