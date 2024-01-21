package spot

import (
  "net/http"
  "strings"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TickersHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *repositories.TickersRepository
}

func NewTickersRouter(apiContext *common.ApiContext) http.Handler {
  h := TickersHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.TickersRepository{
    Rdb: h.ApiContext.Rdb,
    Ctx: h.ApiContext.Ctx,
  }

  r := chi.NewRouter()
  r.Get("/", h.Gets)

  return r
}

func (h *TickersHandler) Gets(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  if r.URL.Query().Get("symbols") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "symbols is empty")
    return
  }

  if r.URL.Query().Get("fields") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "fields is empty")
    return
  }

  symbols := strings.Split(r.URL.Query().Get("symbols"), ",")
  fields := strings.Split(r.URL.Query().Get("fields"), ",")

  tickers := h.Repository.Gets(symbols, fields)

  h.Response.Json(tickers)
}
