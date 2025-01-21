package v1

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/api/v1/dydx"
  "taoniu.local/cryptos/common"
)

func NewDydxRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Use(api.Authenticator)
  r.Mount("/account", dydx.NewAccountRouter(apiContext))
  r.Mount("/tickers", dydx.NewTickersRouter(apiContext))
  r.Mount("/strategies", dydx.NewStrategiesRouter(apiContext))
  r.Mount("/indicators", dydx.NewIndicatorsRouter(apiContext))
  r.Mount("/plans", dydx.NewPlansRouter(apiContext))
  r.Mount("/orders", dydx.NewOrdersRouter(apiContext))
  r.Mount("/positions", dydx.NewPositionsRouter(apiContext))
  r.Mount("/scalping", dydx.NewScalpingRouter(apiContext))
  r.Mount("/tradings", dydx.NewTradingsouter(apiContext))
  r.Mount("/tradingview", dydx.NewTradingViewRouter(apiContext))
  return r
}
