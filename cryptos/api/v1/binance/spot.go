package binance

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api/v1/binance/spot"
)

func NewSpotRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/symbols", spot.NewSymbolsRouter())
	r.Mount("/tickers", spot.NewTickersRouter())
	r.Mount("/klines", spot.NewKlinesRouter())
	r.Mount("/indicators", spot.NewIndicatorsRouter())
	r.Mount("/plans", spot.NewPlansRouter())
	r.Mount("/margin", spot.NewMarginRouter())
	return r
}
