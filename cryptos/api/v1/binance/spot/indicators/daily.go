package indicators

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
)

type DailyHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  //Repository *repositories.DailyRepository
}

func NewDailyRouter(apiContext *common.ApiContext) http.Handler {
  h := DailyHandler{
    ApiContext: apiContext,
  }
  //h.Repository = &repositories.DailyRepository{
  //  Rdb: common.NewRedis(1),
  //  Ctx: context.Background(),
  //}

  r := chi.NewRouter()
  r.Get("/", h.Gets)

  return r
}

func (h *DailyHandler) Gets(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  //h.Response = &api.ResponseHandler{
  //  Writer: w,
  //}
  //
  //if r.URL.Query().Get("symbols") == "" {
  //  h.Response.Error(http.StatusForbidden, 1004, "symbols is empty")
  //  return
  //}
  //
  //if r.URL.Query().Get("fields") == "" {
  //  h.Response.Error(http.StatusForbidden, 1004, "fields is empty")
  //  return
  //}
  //
  //symbols := strings.Split(r.URL.Query().Get("symbols"), ",")
  //fields := strings.Split(r.URL.Query().Get("fields"), ",")
  //
  //indicators := h.Repository.Gets(symbols, fields)
  //
  //h.Response.Json(indicators)
}
