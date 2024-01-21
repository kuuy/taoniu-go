package tradingview

import (
  "net/http"
  "strconv"
  "strings"

  "github.com/go-chi/chi/v5"
  "gorm.io/datatypes"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/tradingview"
)

type AnalysisHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *repositories.AnalysisRepository
}

type AnalysisInfo struct {
  ID        string            `json:"id"`
  Symbol    string            `json:"symbol"`
  Summary   datatypes.JSONMap `json:"summary"`
  Timestamp int64             `json:"timestamp"`
}

func NewAnalysisRouter(apiContext *common.ApiContext) http.Handler {
  h := AnalysisHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.AnalysisRepository{
    Db:  h.ApiContext.Db,
    Rdb: h.ApiContext.Rdb,
    Ctx: h.ApiContext.Ctx,
  }

  r := chi.NewRouter()
  r.Get("/", h.Listings)
  r.Get("/gets", h.Gets)

  return r
}

func (h *AnalysisHandler) Listings(
  w http.ResponseWriter,
  r *http.Request,
) {
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
  }
  pageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
  if pageSize < 1 || pageSize > 100 {
    h.Response.Error(http.StatusForbidden, 1004, "page size not valid")
    return
  }

  conditions := make(map[string]interface{})
  if r.URL.Query().Get("exchange") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "exchange is empty")
    return
  }
  conditions["exchange"] = r.URL.Query().Get("exchange")

  if r.URL.Query().Get("interval") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "interval is empty")
    return
  }
  conditions["interval"] = r.URL.Query().Get("interval")

  total := h.Repository.Count(conditions)
  analysis := h.Repository.Listings(current, pageSize, conditions)
  data := make([]*AnalysisInfo, len(analysis))
  for i, item := range analysis {
    data[i] = &AnalysisInfo{
      ID:        item.ID,
      Symbol:    item.Symbol,
      Summary:   item.Summary,
      Timestamp: item.UpdatedAt.Unix(),
    }
  }

  h.Response.Pagenate(data, total, current, pageSize)
}

func (h *AnalysisHandler) Gets(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  if r.URL.Query().Get("exchange") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "exchange is empty")
    return
  }
  if r.URL.Query().Get("symbols") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "symbols is empty")
    return
  }
  if r.URL.Query().Get("interval") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "interval is empty")
    return
  }

  exchange := r.URL.Query().Get("exchange")
  symbols := strings.Split(r.URL.Query().Get("symbols"), ",")
  interval := r.URL.Query().Get("interval")

  analysis := h.Repository.Gets(exchange, symbols, interval)
  data := make([]*AnalysisInfo, len(analysis))
  for i, item := range analysis {
    data[i] = &AnalysisInfo{
      ID:        item.ID,
      Symbol:    item.Symbol,
      Summary:   item.Summary,
      Timestamp: item.UpdatedAt.Unix(),
    }
  }

  h.Response.Json(data)
}
