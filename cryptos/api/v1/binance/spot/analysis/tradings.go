package analysis

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api/v1/binance/spot/analysis/tradings"
)

func NewTradingsRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/fishers", tradings.NewFishersRouter())
	return r
}
