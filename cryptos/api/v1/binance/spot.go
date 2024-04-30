package binance

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api/v1/binance/spot"
  "taoniu.local/cryptos/common"
)

func NewSpotRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Mount("/symbols", spot.NewSymbolsRouter(apiContext))
  r.Mount("/tickers", spot.NewTickersRouter(apiContext))
  r.Mount("/klines", spot.NewKlinesRouter(apiContext))
  r.Mount("/indicators", spot.NewIndicatorsRouter(apiContext))
  r.Mount("/strategies", spot.NewStrategiesRouter(apiContext))
  r.Mount("/plans", spot.NewPlansRouter(apiContext))
  r.Mount("/orders", spot.NewOrdersRouter(apiContext))
  r.Mount("/positions", spot.NewPositionsRouter(apiContext))
  r.Mount("/scalping", spot.NewScalpingRouter(apiContext))
  r.Mount("/triggers", spot.NewTriggersRouter(apiContext))
  r.Mount("/margin", spot.NewMarginRouter(apiContext))
  r.Mount("/analysis", spot.NewAnalysisRouter(apiContext))
  r.Mount("/tradings", spot.NewTradingsRouter(apiContext))
  r.Mount("/tradingview", spot.NewTradingViewRouter(apiContext))
  return r
}
