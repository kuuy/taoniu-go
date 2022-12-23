package spot

import (
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
	"taoniu.local/cryptos/api"
	repositories "taoniu.local/cryptos/repositories/binance/spot"

	pool "taoniu.local/cryptos/common"
)

type TickersHandler struct {
	Response   *api.ResponseHandler
	Repository *repositories.TickersRepository
}

func NewTickersRouter() http.Handler {
	h := TickersHandler{}
	h.Repository = &repositories.TickersRepository{
		Rdb: pool.NewRedis(),
		Ctx: context.Background(),
	}

	r := chi.NewRouter()
	r.Get("/", h.Gets)

	return r
}

func (h *TickersHandler) Gets(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.Response = &api.ResponseHandler{
		Writer: w,
	}

	if r.URL.Query().Get("symbols") == "" {
		h.Response.Error(http.StatusForbidden, 1004, "symbols is empty")
		return
	}

	if r.URL.Query().Get("fields") == "" {
		h.Response.Error(http.StatusForbidden, 1004, "fields is empty")
		return
	}

	symbols := strings.Split(r.URL.Query().Get("symbols"), ",")
	fields := strings.Split(r.URL.Query().Get("fields"), ",")

	tickers := h.Repository.Gets(symbols, fields)

	h.Response.Json(tickers)
}
