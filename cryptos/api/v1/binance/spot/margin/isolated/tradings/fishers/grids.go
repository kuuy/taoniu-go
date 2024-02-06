package fishers

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings/fishers"
)

type GridsHandler struct {
  ApiContext *common.ApiContext
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

func NewGridsRouter(apiContext *common.ApiContext) http.Handler {
  h := GridsHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.GridsRepository{
    Db: common.NewDB(),
  }

  r := chi.NewRouter()
  r.Get("/pending", h.Pending)

  return r
}

func (h *GridsHandler) Pending(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  data := h.Repository.Pending()

  h.Response.Json(data)
}
