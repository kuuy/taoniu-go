package fishers

import (
  "net/http"
  "strconv"
  "time"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/analysis/tradings/fishers"
)

type GridInfo struct {
  ID          string      `json:"id"`
  Day         string      `json:"day"`
  BuysCount   int         `json:"buys_count"`
  SellsCount  int         `json:"sells_count"`
  BuysAmount  float64     `json:"buys_amount"`
  SellsAmount float64     `json:"sells_amount"`
  Data        interface{} `json:"data"`
}

type GridsHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *repositories.GridsRepository
}

func NewGridsRouter(apiContext *common.ApiContext) http.Handler {
  h := GridsHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.GridsRepository{
    Db:  h.ApiContext.Db,
    Rdb: h.ApiContext.Rdb,
    Ctx: h.ApiContext.Ctx,
  }

  r := chi.NewRouter()
  r.Get("/", h.Listings)
  r.Get("/series", h.Series)

  return r
}

func (h *GridsHandler) Listings(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  var current int
  if !r.URL.Query().Has("current") {
    current = 1
  }
  current, _ = strconv.Atoi(r.URL.Query().Get("current"))
  if current < 1 {
    h.Response.Error(http.StatusForbidden, 1004, "current not valid")
    return
  }

  var pageSize int
  if !r.URL.Query().Has("page_size") {
    pageSize = 50
  } else {
    pageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
  }
  if pageSize < 1 || pageSize > 100 {
    h.Response.Error(http.StatusForbidden, 1004, "page size not valid")
    return
  }

  total := h.Repository.Count()
  grids := h.Repository.Listings(current, pageSize)
  data := make([]*GridInfo, len(grids))
  for i, grid := range grids {
    data[i] = &GridInfo{
      ID:          grid.ID,
      Day:         time.Time(grid.Day).Format("2006-01-02"),
      BuysCount:   grid.BuysCount,
      SellsCount:  grid.SellsCount,
      BuysAmount:  grid.BuysAmount,
      SellsAmount: grid.SellsAmount,
      Data:        grid.Data,
    }
  }

  h.Response.Pagenate(data, total, current, pageSize)
}

func (h *GridsHandler) Series(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  var limit int
  if !r.URL.Query().Has("limit") {
    limit = 15
  } else {
    limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
  }
  if limit < 1 || limit > 100 {
    h.Response.Error(http.StatusForbidden, 1004, "limit not valid")
    return
  }

  series := h.Repository.Series(limit)
  h.Response.Json(series)
}
