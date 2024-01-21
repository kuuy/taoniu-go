package dydx

import (
  "context"
  "net/http"
  "strconv"
  "strings"

  "github.com/go-chi/chi/v5"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type IndicatorsHandler struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Response          *api.ResponseHandler
  Repository        *repositories.IndicatorsRepository
  MarketsRepository *repositories.MarketsRepository
}

func NewIndicatorsRouter(apiContext *common.ApiContext) http.Handler {
  h := IndicatorsHandler{
    Db:  common.NewDB(),
    Rdb: common.NewRedis(),
    Ctx: context.Background(),
  }
  h.Repository = &repositories.IndicatorsRepository{
    Rdb: h.Rdb,
    Ctx: h.Ctx,
  }
  h.MarketsRepository = &repositories.MarketsRepository{
    Db: h.Db,
  }

  r := chi.NewRouter()
  r.Get("/", h.Gets)
  r.Get("/ranking", h.Ranking)

  return r
}

func (h *IndicatorsHandler) Gets(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  q := r.URL.Query()

  if q.Get("symbols") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "symbols is empty")
    return
  }

  if q.Get("interval") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "interval is empty")
    return
  }

  if q.Get("fields") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "fields is empty")
    return
  }

  symbols := strings.Split(q.Get("symbols"), ",")
  interval := q.Get("interval")
  fields := strings.Split(q.Get("fields"), ",")

  indicators := h.Repository.Gets(symbols, interval, fields)

  h.Response.Json(indicators)
}

func (h *IndicatorsHandler) Ranking(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  q := r.URL.Query()

  if q.Get("interval") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "interval is empty")
    return
  }

  if q.Get("fields") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "fields is empty")
    return
  }

  if q.Get("sort") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "sort is empty")
    return
  }

  var symbols []string
  if q.Get("symbols") == "" {
    symbols = h.MarketsRepository.Symbols()
  } else {
    symbols = strings.Split(q.Get("symbols"), ",")
  }
  interval := q.Get("interval")
  fields := strings.Split(q.Get("fields"), ",")

  sort := strings.Split(q.Get("sort"), ",")
  sortField := sort[0]
  sortType, _ := strconv.Atoi(sort[1])

  var current int
  if !q.Has("current") {
    current = 1
  }
  current, _ = strconv.Atoi(q.Get("current"))
  if current < 1 {
    h.Response.Error(http.StatusForbidden, 1004, "current not valid")
    return
  }

  var pageSize int
  if !q.Has("page_size") {
    pageSize = 50
  } else {
    pageSize, _ = strconv.Atoi(q.Get("page_size"))
  }
  if pageSize < 1 || pageSize > 100 {
    h.Response.Error(http.StatusForbidden, 1004, "page size not valid")
    return
  }

  result := h.Repository.Ranking(symbols, interval, fields, sortField, sortType, current, pageSize)

  h.Response.Pagenate(result.Data, int64(result.Total), current, pageSize)
}