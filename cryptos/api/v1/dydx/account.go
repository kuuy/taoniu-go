package dydx

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type AccountHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *repositories.AccountRepository
}

func NewAccountRouter(apiContext *common.ApiContext) http.Handler {
  h := AccountHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.AccountRepository{
    Rdb: h.ApiContext.Rdb,
    Ctx: h.ApiContext.Ctx,
  }

  r := chi.NewRouter()
  r.Get("/", h.Balance)
  return r
}

func (h *AccountHandler) Balance(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  data, _ := h.Repository.Balance()

  h.Response.Json(data)
}
