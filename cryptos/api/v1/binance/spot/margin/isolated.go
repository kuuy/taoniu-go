package margin

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api/v1/binance/spot/margin/isolated"
)

func NewIsolatedRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/tradings", isolated.NewTradingsRouter())
	return r
}
