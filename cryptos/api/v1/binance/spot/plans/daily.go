package plans

import (
	"context"
	"net/http"
	"strconv"
	"taoniu.local/cryptos/common"
	"time"

	"github.com/go-chi/chi/v5"

	"taoniu.local/cryptos/api"
	repositories "taoniu.local/cryptos/repositories/binance/spot/plans"
)

type DailyHandler struct {
	Response   *api.ResponseHandler
	Repository *repositories.DailyRepository
}

type DailyInfo struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Side      int       `json:"side"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

func NewDailyRouter() http.Handler {
	h := DailyHandler{}
	h.Repository = &repositories.DailyRepository{
		Db:  common.NewDB(),
		Rdb: common.NewRedis(),
		Ctx: context.Background(),
	}

	r := chi.NewRouter()
	r.Get("/", h.Listings)

	return r
}

func (h *DailyHandler) Listings(
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
	}
	pageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		h.Response.Error(http.StatusForbidden, 1004, "page size not valid")
		return
	}

	total := h.Repository.Count()
	plans := h.Repository.Listings(current, pageSize)
	data := make([]*DailyInfo, len(plans))
	for i, plan := range plans {
		data[i] = &DailyInfo{
			ID:        plan.ID,
			Symbol:    plan.Symbol,
			Side:      int(plan.Side),
			Price:     plan.Price,
			Quantity:  plan.Quantity,
			Amount:    plan.Amount,
			CreatedAt: plan.CreatedAt,
		}
	}

	h.Response.Pagenate(data, total, current, pageSize)
}
