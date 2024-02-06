package futures

import (
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type PlansHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *repositories.PlansRepository
}

type PlansInfo struct {
  ID        string  `json:"id"`
  Symbol    string  `json:"symbol"`
  Side      int     `json:"side"`
  Price     float64 `json:"price"`
  Quantity  float64 `json:"quantity"`
  Amount    float64 `json:"amount"`
  Timestamp int64   `json:"timestamp"`
  Status    int     `json:"status"`
}

func NewPlansRouter(apiContext *common.ApiContext) http.Handler {
  h := PlansHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.PlansRepository{
    Db: h.ApiContext.Db,
  }

  r := chi.NewRouter()
  r.Get("/", h.Listings)

  return r
}

func (h *PlansHandler) Listings(
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
  if q.Get("side") != "" {
    conditions["side"], _ = strconv.Atoi(q.Get("side"))
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
  plans := h.Repository.Listings(conditions, current, pageSize)
  data := make([]*PlansInfo, len(plans))
  for i, plan := range plans {
    data[i] = &PlansInfo{
      ID:        plan.ID,
      Symbol:    plan.Symbol,
      Side:      plan.Side,
      Price:     plan.Price,
      Quantity:  plan.Quantity,
      Amount:    plan.Amount,
      Status:    plan.Status,
      Timestamp: plan.Timestamp,
    }
  }

  h.Response.Pagenate(data, total, current, pageSize)
}
