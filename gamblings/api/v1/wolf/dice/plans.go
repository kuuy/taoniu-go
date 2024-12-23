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

type PlansHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Response   *api.ResponseHandler
	Repository *repositories.PlansRepository
}

type PlanShowInfo struct {
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

func NewPlansRouter() http.Handler {
	h := PlansHandler{
		Rdb: common.NewRedis(),
		Ctx: context.Background(),
	}
	h.Repository = &repositories.PlansRepository{
		Db:  common.NewDB(),
		Rdb: h.Rdb,
		Ctx: h.Ctx,
	}

	r := chi.NewRouter()
	r.Get("/show", h.Show)

	return r
}

func (h *PlansHandler) Show(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.Response = &api.ResponseHandler{
		Writer: w,
	}
	plan, err := h.Repository.Show()
	if err != nil {
		h.Response.Error(http.StatusForbidden, 1000, err.Error())
		return
	}

	result := &PlanShowInfo{
		Currency:      plan.Currency,
		Balance:       plan.Balance,
		Invest:        plan.Invest,
		Profit:        plan.Profit,
		BestProfit:    plan.BestProfit,
		WinAmount:     plan.WinAmount,
		LossAmount:    plan.LossAmount,
		StartBalance:  plan.StartBalance,
		TargetBalance: plan.TargetBalance,
		StopBalance:   plan.StopBalance,
	}

	h.Response.Json(result)
}
