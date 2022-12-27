package currencies

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api"
	"taoniu.local/cryptos/common"
	"taoniu.local/cryptos/repositories"
)

type AboutHandler struct {
	Response   *api.ResponseHandler
	Repository *repositories.CurrenciesRepository
}

func NewAboutRouter() http.Handler {
	h := AboutHandler{}
	h.Repository = &repositories.CurrenciesRepository{
		Db: common.NewDB(),
	}

	r := chi.NewRouter()
	r.Get("/", h.Get)

	return r
}

func (h *AboutHandler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.Response = &api.ResponseHandler{
		Writer: w,
	}

	symbol := chi.URLParam(r, "symbol")
	about, err := h.Repository.About(symbol)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	h.Response.Json(about)
}
