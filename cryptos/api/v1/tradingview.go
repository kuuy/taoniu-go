package v1

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api"
	"taoniu.local/cryptos/api/v1/tradingview"
)

func NewTradingviewRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(api.Authenticator)
	r.Mount("/analysis", tradingview.NewAnalysisRouter())

	return r
}
