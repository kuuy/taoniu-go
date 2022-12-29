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
	r.Mount("/summary/{exchange:[A-Z0-9]{1,20}},{symbol:[A-Z0-9]{1,20}},{interval:[a-zA-Z0-9]{2}}", tradingview.NewSummaryRouter())

	return r
}
