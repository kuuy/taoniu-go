package futures

import (
  "net/http"
  "strconv"
  "strings"

  "github.com/go-chi/chi/v5"
  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/repositories"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
)

type OrdersHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *futuresRepositories.OrdersRepository
}

func NewOrdersRouter(apiContext *common.ApiContext) http.Handler {
  h := OrdersHandler{
    ApiContext: apiContext,
  }
  h.Response = &api.ResponseHandler{}
  h.Response.JweRepository = &repositories.JweRepository{}
  h.Repository = &futuresRepositories.OrdersRepository{
    Db:  h.ApiContext.Db,
    Rdb: h.ApiContext.Rdb,
    Ctx: h.ApiContext.Ctx,
  }

  r := chi.NewRouter()
  r.Get("/", h.Listings)
  r.Get("/{id:[a-z0-9]{20}}", h.Cancel)
  return r
}

func (h *OrdersHandler) Listings(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response.Writer = w

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

  conditions := make(map[string]interface{})

  if r.URL.Query().Get("symbols") != "" {
    conditions["symbols"] = strings.Split(r.URL.Query().Get("symbols"), ",")
  }

  if r.URL.Query().Get("position_side") != "" {
    conditions["position_side"] = r.URL.Query().Get("position_side")
  }

  if r.URL.Query().Get("status") != "" {
    conditions["status"] = r.URL.Query().Get("status")
  }

  total := h.Repository.Count(conditions)
  orders := h.Repository.Listings(conditions, current, pageSize)
  data := make([]*OrderInfo, len(orders))
  for i, order := range orders {
    data[i] = &OrderInfo{
      ID:           order.ID,
      Symbol:       order.Symbol,
      OrderId:      order.OrderId,
      PositionSide: order.PositionSide,
      Side:         order.Side,
      Price:        order.Price,
      Quantity:     order.Quantity,
      OpenTime:     order.OpenTime,
      UpdateTime:   order.UpdateTime,
      ReduceOnly:   order.ReduceOnly,
      Status:       order.Status,
    }
  }

  h.Response.Pagenate(data, total, current, pageSize)
}

func (h *OrdersHandler) Cancel(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response.Writer = w

  //id := chi.URLParam(r, "id")
  //err := h.Repository.Cancel(id)
  //if err != nil {
  //	h.Response.Error(http.StatusForbidden, 1004, err.Error())
  //	return
  //}

  h.Response.Json(nil)
}
