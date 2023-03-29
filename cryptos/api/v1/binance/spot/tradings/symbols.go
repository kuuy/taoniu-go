package tradings

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api"
	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type SymbolsHandler struct {
	Response   *api.ResponseHandler
	Repository *repositories.SymbolsRepository
}

type SymbolInfo struct {
	Symbol     string `json:"symbol"`
	BaseAsset  string `json:"base_asset"`
	QuoteAsset string `json:"quote_asset"`
}

func NewSymbolsRouter() http.Handler {
	h := SymbolsHandler{}
	h.Repository = &repositories.SymbolsRepository{
		Db: common.NewDB(),
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	symbols := h.Repository.Scan()
	h.Response.Json(symbols)
}
