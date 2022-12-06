package spot

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api/v1/binance/spot/margin"
)

func NewMarginRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/orders", margin.NewOrdersRouter())
	r.Mount("/isolated", margin.NewIsolatedRouter())
	return r
}
