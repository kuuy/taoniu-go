package spot

import (
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type ScalpingHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *repositories.ScalpingRepository
}

type ScalpingInfo struct {
  ID          string  `json:"id"`
  Symbol      string  `json:"symbol"`
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
  scalpings := h.Repository.Listings(conditions, current, pageSize)
  data := make([]*ScalpingInfo, len(scalpings))
  for i, scalping := range scalpings {
    data[i] = &ScalpingInfo{
      ID:          scalping.ID,
      Symbol:      scalping.Symbol,
      Capital:     scalping.Capital,
      Price:       scalping.Price,
      TakePrice:   scalping.TakePrice,
      StopPrice:   scalping.StopPrice,
      TakeOrderId: scalping.TakeOrderId,
      StopOrderId: scalping.StopOrderId,
      Profit:      scalping.Profit,
      Timestamp:   scalping.Timestamp,
      Status:      scalping.Status,
      ExpiredAt:   scalping.ExpiredAt.Unix(),
      CreatedAt:   scalping.CreatedAt.Unix(),
    }
  }

  h.Response.Pagenate(data, total, current, pageSize)
}
