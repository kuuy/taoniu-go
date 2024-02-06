package spot

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type SymbolsHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *repositories.SymbolsRepository
}

type SymbolInfo struct {
  Symbol     string `json:"symbol"`
  BaseAsset  string `json:"base_asset"`
  QuoteAsset string `json:"quote_asset"`
}

func NewSymbolsRouter(apiContext *common.ApiContext) http.Handler {
  h := SymbolsHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.SymbolsRepository{
    Db: h.ApiContext.Db,
  }

  r := chi.NewRouter()
  r.Get("/{symbol:[A-Z0-9]{1,20}}", h.Get)
  return r
}

func (h *SymbolsHandler) Get(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  symbol := chi.URLParam(r, "symbol")
  entity, err := h.Repository.Get(symbol)
  if err != nil {
    http.Error(w, http.StatusText(404), http.StatusNotFound)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  result := &SymbolInfo{
    Symbol:     entity.Symbol,
    BaseAsset:  entity.BaseAsset,
    QuoteAsset: entity.QuoteAsset,
  }

  h.Response.Json(result)
}
