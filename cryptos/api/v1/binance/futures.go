package binance

import (
  "github.com/go-chi/chi/v5"
  "net/http"

  "taoniu.local/cryptos/api/v1/binance/futures"
  "taoniu.local/cryptos/common"
)

func NewFuturesRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/tickers", futures.NewTickersRouter(apiContext))
  r.Mount("/indicators", futures.NewIndicatorsRouter(apiContext))
  r.Mount("/strategies", futures.NewStrategiesRouter(apiContext))
  r.Mount("/plans", futures.NewPlansRouter(apiContext))
  r.Mount("/orders", futures.NewOrdersRouter(apiContext))
  r.Mount("/positions", futures.NewPositionsRouter(apiContext))
  r.Mount("/gambling", futures.NewGamblingRouter(apiContext))
  r.Mount("/scalping", futures.NewScalpingRouter(apiContext))
  r.Mount("/analysis", futures.NewAnalysisRouter(apiContext))
  r.Mount("/tradings", futures.NewTradingsRouter(apiContext))
  r.Mount("/tradingview", futures.NewTradingViewRouter(apiContext))
  return r
}
