package templates

import (
  "context"
  "github.com/go-chi/chi/v5"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"
  "net/http"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
)

type DrawingHandler struct {
  Db       *gorm.DB
  Rdb      *redis.Client
  Ctx      context.Context
  Response *api.ResponseHandler
}

func NewDrawingRouter(apiContext *common.ApiContext) http.Handler {
  h := DrawingHandler{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }

  r := chi.NewRouter()
  r.Get("/", h.Gets)

  return r
}

func (h *DrawingHandler) Gets(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  h.Response.Out("")
}