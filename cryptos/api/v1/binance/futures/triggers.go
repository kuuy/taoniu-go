package futures

import (
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type TriggersHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *repositories.TriggersRepository
}

type TriggersInfo struct {
  ID          string  `json:"id"`
  Symbol      string  `json:"symbol"`
  Side        int     `json:"side"`
  Capital     float64 `json:"capital"`
  Price       float64 `json:"price"`
  TakePrice   float64 `json:"take_price"`
  StopPrice   float64 `json:"stop_price"`
  TakeOrderId int64   `json:"take_order_id"`
  StopOrderId int64   `json:"stop_order_id"`
  Profit      float64 `json:"profit"`
  Timestamp   int64   `json:"timestamp"`
  Status      int     `json:"status"`
  ExpiredAt   int64   `json:"expired_at"`
  CreatedAt   int64   `json:"created_at"`
}

func NewTriggersRouter(apiContext *common.ApiContext) http.Handler {
  h := TriggersHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.TriggersRepository{
    Db: h.ApiContext.Db,
  }

  r := chi.NewRouter()
  r.Get("/", h.Listings)

  return r
}

func (h *TriggersHandler) Listings(
  w http.ResponseWriter,
  r *http.Request,
) {
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
  triggers := h.Repository.Listings(conditions, current, pageSize)
  data := make([]*TriggersInfo, len(triggers))
  for i, trigger := range triggers {
    data[i] = &TriggersInfo{
      ID:          trigger.ID,
      Symbol:      trigger.Symbol,
      Side:        trigger.Side,
      Capital:     trigger.Capital,
      Price:       trigger.Price,
      TakePrice:   trigger.TakePrice,
      StopPrice:   trigger.StopPrice,
      TakeOrderId: trigger.TakeOrderId,
      StopOrderId: trigger.StopOrderId,
      Profit:      trigger.Profit,
      Timestamp:   trigger.Timestamp,
      Status:      trigger.Status,
      ExpiredAt:   trigger.ExpiredAt.Unix(),
      CreatedAt:   trigger.CreatedAt.Unix(),
    }
  }

  h.Response.Pagenate(data, total, current, pageSize)
}
