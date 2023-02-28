package spot

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api/v1/binance/spot/analysis"
)

func NewAnalysisRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/margin", analysis.NewMarginRouter())
	return r
}
