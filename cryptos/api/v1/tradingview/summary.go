package tradingview

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api"
	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/tradingview"
)

type SummaryHandler struct {
	Response   *api.ResponseHandler
	Repository *repositories.AnalysisRepository
}

func NewSummaryRouter() http.Handler {
	h := SummaryHandler{}
	h.Repository = &repositories.AnalysisRepository{
		Db: common.NewDB(),
	}

	r := chi.NewRouter()
	r.Get("/", h.Get)

	return r
}

func (h *SummaryHandler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.Response = &api.ResponseHandler{
		Writer: w,
	}

	exchange := chi.URLParam(r, "exchange")
	symbol := chi.URLParam(r, "symbol")
	interval := chi.URLParam(r, "interval")
	summary, err := h.Repository.Summary(exchange, symbol, interval)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	h.Response.Json(summary)
}
