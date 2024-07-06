package isolated

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"

  repositories "taoniu.local/cryptos/repositories/binance/margin/isolated"
)

type TradingsHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *repositories.TradingsRepository
}

func NewTradingsRouter(apiContext *common.ApiContext) http.Handler {
  h := TradingsHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.TradingsRepository{
    Db: h.ApiContext.Db,
  }

  r := chi.NewRouter()
  r.Get("/scan", h.Scan)
  return r
}

func (h *TradingsHandler) Scan(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }
  symbols := h.Repository.Scan()
  h.Response.Json(symbols)
}
