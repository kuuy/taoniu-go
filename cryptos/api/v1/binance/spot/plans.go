package spot

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api/v1/binance/spot/plans"
)

func NewPlansRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/daily", plans.NewDailyRouter())
	return r
}
