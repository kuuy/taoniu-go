package spot

import (
  "net/http"
  "strconv"
  "strings"

  "github.com/go-chi/chi/v5"
  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/repositories"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
)

type IndicatorsHandler struct {
  ApiContext        *common.ApiContext
  Response          *api.ResponseHandler
  Repository        *spotRepositories.IndicatorsRepository
  SymbolsRepository *spotRepositories.SymbolsRepository
}

func NewIndicatorsRouter(apiContext *common.ApiContext) http.Handler {
  h := IndicatorsHandler{
    ApiContext: apiContext,
  }
  h.Response = &api.ResponseHandler{}
  h.Response.JweRepository = &repositories.JweRepository{}
  h.Repository = &spotRepositories.IndicatorsRepository{
    Rdb: h.ApiContext.Rdb,
    Ctx: h.ApiContext.Ctx,
  }
  h.SymbolsRepository = &spotRepositories.SymbolsRepository{
    Db: h.ApiContext.Db,
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
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response.Writer = w

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
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response.Writer = w

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
    symbols = h.SymbolsRepository.Symbols()
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
