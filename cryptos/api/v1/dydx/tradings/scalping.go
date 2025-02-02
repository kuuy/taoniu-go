package tradings

import (
  "net/http"
  "strconv"
  "strings"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx/tradings"
)

type ScalpingHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *repositories.ScalpingRepository
}

func NewScalpingRouter(apiContext *common.ApiContext) http.Handler {
  h := ScalpingHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.ScalpingRepository{
    Db: h.ApiContext.Db,
  }

  r := chi.NewRouter()
  r.Get("/", h.Listings)

  return r
}

func (h *ScalpingHandler) Listings(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  q := r.URL.Query()
  conditions := make(map[string]interface{})
  if q.Get("symbol") != "" {
    conditions["symbol"] = q.Get("symbol")
  }
  if q.Get("status") != "" {
    conditions["status"] = []int{}
    for _, status := range strings.Split(q.Get("status"), ",") {
      status, _ := strconv.Atoi(status)
      conditions["status"] = append(conditions["status"].([]int), status)
    }
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
  tradings := h.Repository.Listings(conditions, current, pageSize)
  data := make([]*ScalpingTradingInfo, len(tradings))
  for i, trading := range tradings {
    data[i] = &ScalpingTradingInfo{
      ID:           trading.ID,
      Symbol:       trading.Symbol,
      ScalpingId:   trading.ScalpingId,
      PlanId:       trading.PlanId,
      BuyPrice:     trading.BuyPrice,
      SellPrice:    trading.SellPrice,
      BuyQuantity:  trading.BuyQuantity,
      SellQuantity: trading.SellQuantity,
      BuyOrderId:   trading.BuyOrderId,
      SellOrderId:  trading.SellOrderId,
      Status:       trading.Status,
      CreatedAt:    trading.CreatedAt.Unix(),
      UpdatedAt:    trading.UpdatedAt.Unix(),
    }
  }

  h.Response.Pagenate(data, total, current, pageSize)
}
