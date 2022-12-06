package tradings

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"taoniu.local/cryptos/api"
	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
)

type GridsHandler struct {
	Response   *api.ResponseHandler
	Repository *repositories.GridsRepository
}

type GridInfo struct {
	ID              string  `json:"id"`
	Symbol          string  `json:"symbol"`
	BuyPrice        float64 `json:"buy_price"`
	SellPrice       float64 `json:"sell_price"`
	Status          int     `json:"status"`
	Timestamp       int64   `json:"timestamp"`
	TimestampFormat string  `json:"timestamp_fmt"`
}

func NewGridsRouter() http.Handler {
	h := GridsHandler{}
	h.Repository = &repositories.GridsRepository{
		Db:  common.NewDB(),
		Rdb: common.NewRedis(),
		Ctx: context.Background(),
	}

	r := chi.NewRouter()
	r.Get("/", h.Listings)

	return r
}

func (h *GridsHandler) Listings(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.Response = &api.ResponseHandler{
		Writer: w,
	}

	var current int
	if !r.URL.Query().Has("current") {
		current = 1
	}
	current, _ = strconv.Atoi(r.URL.Query().Get("current"))
	if current < 1 {
		h.Response.Error(http.StatusForbidden, 1004, "current not valid")
		return
	}

	var pageSize int
	if !r.URL.Query().Has("page_size") {
		pageSize = 50
	} else {
		pageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
	}
	if pageSize < 1 || pageSize > 100 {
		h.Response.Error(http.StatusForbidden, 1004, "page size not valid")
		return
	}

	total := h.Repository.Count()
	tradings := h.Repository.Listings(current, pageSize)
	data := make([]*GridInfo, len(tradings))
	for i, trade := range tradings {
		data[i] = &GridInfo{
			ID:              trade.ID,
			Symbol:          trade.Symbol,
			BuyPrice:        trade.BuyPrice,
			SellPrice:       trade.SellPrice,
			Status:          trade.Status,
			Timestamp:       trade.CreatedAt.Unix(),
			TimestampFormat: common.FormatDatetime(trade.CreatedAt),
		}
	}

	h.Response.Pagenate(data, total, current, pageSize)
}
