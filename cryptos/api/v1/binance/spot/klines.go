package spot

import (
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"taoniu.local/cryptos/api"
	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
	"time"
)

type KlinesHandler struct {
	Response   *api.ResponseHandler
	Repository *repositories.KlinesRepository
}

func NewKlinesRouter() http.Handler {
	h := KlinesHandler{}
	h.Repository = &repositories.KlinesRepository{
		Db:  common.NewDB(),
		Rdb: common.NewRedis(),
		Ctx: context.Background(),
	}

	r := chi.NewRouter()
	r.Get("/", h.Series)

	return r
}

func (h *KlinesHandler) Series(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.Response = &api.ResponseHandler{
		Writer: w,
	}

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		h.Response.Error(http.StatusForbidden, 1004, "symbol is empty")
		return
	}

	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "1d"
	}

	var timestamp int64
	if !r.URL.Query().Has("timestamp") {
		timestamp = time.Now().UnixMicro()
	} else {
		timestamp, _ = strconv.ParseInt(r.URL.Query().Get("timestamp"), 10, 64)
	}

	var limit int
	if !r.URL.Query().Has("limit") {
		limit = 50
	} else {
		limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	}
	if limit < 1 || limit > 100 {
		h.Response.Error(http.StatusForbidden, 1004, "limit not valid")
		return
	}

	series := h.Repository.Series(symbol, interval, timestamp, limit)

	h.Response.Json(series)
}
