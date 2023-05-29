package indicators

import (
  "context"
  "net/http"
  "strings"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/indicators"
)

type DailyHandler struct {
  Response   *api.ResponseHandler
  Repository *repositories.DailyRepository
}

func NewDailyRouter() http.Handler {
  h := DailyHandler{}
  h.Repository = &repositories.DailyRepository{
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }

  r := chi.NewRouter()
  r.Get("/", h.Gets)

  return r
}

func (h *DailyHandler) Gets(
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

  indicators := h.Repository.Gets(symbols, fields)

  h.Response.Json(indicators)
}
