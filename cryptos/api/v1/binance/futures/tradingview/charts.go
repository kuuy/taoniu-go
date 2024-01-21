package tradingview

import (
  "context"
  "github.com/go-chi/chi/v5"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"
  "net/http"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
)

type ChartsHandler struct {
  Db       *gorm.DB
  Rdb      *redis.Client
  Ctx      context.Context
  Response *api.ResponseHandler
}

type ChartInfo struct {
  ID         string `json:"id"`
  Symbol     string `json:"symbol"`
  Name       string `json:"name"`
  Resolution string `json:"resolution"`
  Timestamp  int64  `json:"timestamp"`
}

func NewChartsRouter(apiContext *common.ApiContext) http.Handler {
  h := ChartsHandler{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }

  r := chi.NewRouter()
  r.Get("/", h.Gets)

  return r
}

func (h *ChartsHandler) Gets(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  var charts []*ChartInfo

  h.Response.Out(map[string]interface{}{
    "status": "ok",
    "data":   charts,
  })
}
