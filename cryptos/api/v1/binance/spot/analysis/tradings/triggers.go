package tradings

import (
  "fmt"
  "net/http"
  "strconv"
  "time"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/repositories"
  analysisRepositories "taoniu.local/cryptos/repositories/binance/spot/analysis/tradings"
)

type TriggersHandler struct {
  ApiContext         *common.ApiContext
  Response           *api.ResponseHandler
  AnalysisRepository *analysisRepositories.TriggersRepository
}

func NewTriggersRouter(apiContext *common.ApiContext) http.Handler {
  h := TriggersHandler{
    ApiContext: apiContext,
  }
  h.Response = &api.ResponseHandler{}
  h.Response.JweRepository = &repositories.JweRepository{}
  h.AnalysisRepository = &analysisRepositories.TriggersRepository{
    Db:  h.ApiContext.Db,
    Rdb: h.ApiContext.Rdb,
    Ctx: h.ApiContext.Ctx,
  }

  r := chi.NewRouter()
  r.Get("/", h.Listings)
  r.Get("/series", h.Series)

  return r
}

func (h *TriggersHandler) Listings(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response.Writer = w

  var current int
  if !r.URL.Query().Has("current") {
    current = 1
  } else {
    current, _ = strconv.Atoi(r.URL.Query().Get("current"))
  }
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

  conditions := map[string]interface{}{}

  total := h.AnalysisRepository.Count(conditions)
  tradings := h.AnalysisRepository.Listings(conditions, current, pageSize)
  data := make([]*TriggerInfo, len(tradings))
  for i, trading := range tradings {
    data[i] = &TriggerInfo{
      ID:             trading.ID,
      Day:            time.Time(trading.Day).Format("2006-01-02"),
      BuysCount:      trading.BuysCount,
      SellsCount:     trading.SellsCount,
      BuysAmount:     fmt.Sprintf("%.2f", trading.BuysAmount),
      SellsAmount:    fmt.Sprintf("%.2f", trading.SellsAmount),
      Profit:         fmt.Sprintf("%.2f", trading.Profit),
      AdditiveProfit: fmt.Sprintf("%.2f", trading.AdditiveProfit),
    }
  }

  h.Response.Pagenate(data, total, current, pageSize)
}

func (h *TriggersHandler) Series(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response.Writer = w

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

  series := h.AnalysisRepository.Series(limit)
  h.Response.Json(series)
}
