package v1

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/api/v1/currencies"
  "taoniu.local/cryptos/common"
)

func NewCurrenciesRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Use(api.Authenticator)
  r.Mount("/about/{symbol:[A-Z0-9]{1,20}}", currencies.NewAboutRouter(apiContext))

  return r
}
