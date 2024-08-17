package spot

import (
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/repositories"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
)

type StrategiesHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *spotRepositories.StrategiesRepository
}

func NewStrategiesRouter(apiContext *common.ApiContext) http.Handler {
  h := StrategiesHandler{
    ApiContext: apiContext,
  }
  h.Response = &api.ResponseHandler{}
  h.Response.JweRepository = &repositories.JweRepository{}
  h.Repository = &spotRepositories.StrategiesRepository{
    Db: h.ApiContext.Db,
  }

  r := chi.NewRouter()
  r.Get("/", h.Listings)

  return r
}

func (h *StrategiesHandler) Listings(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response.Writer = w

  q := r.URL.Query()
  conditions := make(map[string]interface{})
  if q.Get("symbol") != "" {
    conditions["symbol"] = q.Get("symbol")
  }
  if q.Get("signal") != "" {
    conditions["signal"], _ = strconv.Atoi(q.Get("signal"))
  }

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

  total := h.Repository.Count(conditions)
  strategies := h.Repository.Listings(conditions, current, pageSize)
  data := make([]*StrategiesInfo, len(strategies))
  for i, strategy := range strategies {
    data[i] = &StrategiesInfo{
      ID:        strategy.ID,
      Symbol:    strategy.Symbol,
      Indicator: strategy.Indicator,
      Signal:    strategy.Signal,
      Price:     strategy.Price,
      Timestamp: strategy.Timestamp,
    }
  }

  h.Response.Pagenate(data, total, current, pageSize)
}
