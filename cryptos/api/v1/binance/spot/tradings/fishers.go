package tradings

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api/v1/binance/spot/tradings/fishers"
)

func NewFishersRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/grids", fishers.NewGridsRouter())
	return r
}
