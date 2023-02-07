package isolated

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"net/http"
	"taoniu.local/cryptos/api"
	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type SymbolsHandler struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
	Response   *api.ResponseHandler
	Repository *repositories.SymbolsRepository
}

func NewSymbolsRouter() http.Handler {
	h := SymbolsHandler{
		Db: common.NewDB(),
	}
	h.Repository = &repositories.SymbolsRepository{
		Db: h.Db,
	}

	r := chi.NewRouter()
	r.Get("/scan", h.Scan)
	return r
}

func (h *SymbolsHandler) Scan(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.Response = &api.ResponseHandler{
		Writer: w,
	}
	symbols := h.Repository.Scan()
	h.Response.Json(symbols)
}
