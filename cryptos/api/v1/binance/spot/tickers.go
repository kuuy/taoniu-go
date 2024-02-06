package spot

import (
  "net/http"
  "strconv"
  "strings"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TickersHandler struct {
  ApiContext        *common.ApiContext
  Response          *api.ResponseHandler
  Repository        *repositories.TickersRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewTickersRouter(apiContext *common.ApiContext) http.Handler {
  h := TickersHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.TickersRepository{
    Rdb: h.ApiContext.Rdb,
    Ctx: h.ApiContext.Ctx,
  }
  h.SymbolsRepository = &repositories.SymbolsRepository{
    Db: h.ApiContext.Db,
  }

  r := chi.NewRouter()
  r.Get("/", h.Gets)
  r.Get("/ranking", h.Ranking)

  return r
}

func (h *TickersHandler) Gets(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

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

func (h *TickersHandler) Ranking(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  q := r.URL.Query()

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

  result := h.Repository.Ranking(symbols, fields, sortField, sortType, current, pageSize)

  h.Response.Pagenate(result.Data, int64(result.Total), current, pageSize)
}
