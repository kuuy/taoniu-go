package spot

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api/v1/binance/spot/indicators"
)

func NewIndicatorsRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/daily", indicators.NewDailyRouter())
	return r
}
