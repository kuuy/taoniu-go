package dydx

import (
  "context"
  "github.com/go-chi/chi/v5"
  "github.com/go-redis/redis/v8"
  "net/http"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type AccountHandler struct {
  Rdb        *redis.Client
  Ctx        context.Context
  Response   *api.ResponseHandler
  Repository *repositories.AccountRepository
}

func NewAccountRouter() http.Handler {
  h := AccountHandler{
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.AccountRepository{
    Rdb: h.Rdb,
    Ctx: h.Ctx,
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
