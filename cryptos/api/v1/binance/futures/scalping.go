package futures

import (
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/repositories"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
)

type ScalpingHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *futuresRepositories.ScalpingRepository
}

func NewScalpingRouter(apiContext *common.ApiContext) http.Handler {
  h := ScalpingHandler{
    ApiContext: apiContext,
  }
  h.Response = &api.ResponseHandler{}
  h.Response.JweRepository = &repositories.JweRepository{}
  h.Repository = &futuresRepositories.ScalpingRepository{
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

  h.Response.Writer = w

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
  scalpings := h.Repository.Listings(conditions, current, pageSize)
  data := make([]*ScalpingInfo, len(scalpings))
  for i, scalping := range scalpings {
    data[i] = &ScalpingInfo{
      ID:          scalping.ID,
      Symbol:      scalping.Symbol,
      Side:        scalping.Side,
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
