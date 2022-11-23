package dice

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"net/http"
	"taoniu.local/gamblings/api"
	"taoniu.local/gamblings/common"
	repositories "taoniu.local/gamblings/repositories/wolf/dice"
)

type MultipleHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Response   *api.ResponseHandler
	Repository *repositories.MultipleRepository
}

type MultipleShowInfo struct {
	Currency      string  `json:"currency"`
	Balance       float64 `json:"balance"`
	Invest        float64 `json:"invest"`
	Profit        float64 `json:"profit"`
	BestProfit    float64 `json:"best_profit"`
	WinAmount     float64 `json:"win_amount"`
	LossAmount    float64 `json:"loss_amount"`
	StartBalance  float64 `json:"start_balance"`
	TargetBalance float64 `json:"target_balance"`
	StopBalance   float64 `json:"stop_balance"`
}

func NewMultipleRouter() http.Handler {
	h := MultipleHandler{
		Rdb: common.NewRedis(),
		Ctx: context.Background(),
	}
	h.Repository = &repositories.MultipleRepository{
		Db:  common.NewDB(),
		Rdb: h.Rdb,
		Ctx: h.Ctx,
	}

	r := chi.NewRouter()
	r.Get("/show", h.Show)

	return r
}

func (h *MultipleHandler) Show(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.Response = &api.ResponseHandler{
		Writer: w,
	}
	multiple, err := h.Repository.Show()
	if err != nil {
		h.Response.Error(http.StatusForbidden, 1000, err.Error())
		return
	}

	result := &MultipleShowInfo{
		Currency:      multiple.Currency,
		Balance:       multiple.Balance,
		Invest:        multiple.Invest,
		Profit:        multiple.Profit,
		BestProfit:    multiple.BestProfit,
		WinAmount:     multiple.WinAmount,
		LossAmount:    multiple.LossAmount,
		StartBalance:  multiple.StartBalance,
		TargetBalance: multiple.TargetBalance,
		StopBalance:   multiple.StopBalance,
	}

	h.Response.Json(result)
}
