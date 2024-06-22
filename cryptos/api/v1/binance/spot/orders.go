package spot

import (
  "net/http"
  "strconv"
  "strings"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type OrdersHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *repositories.OrdersRepository
}

type OrderInfo struct {
  ID              string  `json:"id"`
  Symbol          string  `json:"symbol"`
  OrderId         int64   `json:"order_id"`
  Type            string  `json:"type"`
  Side            string  `json:"side"`
  Price           float64 `json:"price"`
  Quantity        float64 `json:"quantity"`
  OpenTime        int64   `json:"open_time"`
  UpdateTime      int64   `json:"update_time"`
  Status          string  `json:"status"`
  Timestamp       int64   `json:"timestamp"`
  TimestampFormat string  `json:"timestamp_fmt"`
}

func NewOrdersRouter(apiContext *common.ApiContext) http.Handler {
  h := OrdersHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.OrdersRepository{
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

  conditions := make(map[string]interface{})

  if r.URL.Query().Get("symbols") != "" {
    conditions["symbols"] = strings.Split(r.URL.Query().Get("symbols"), ",")
  }

  total := h.Repository.Count(conditions)
  orders := h.Repository.Listings(conditions, current, pageSize)
  data := make([]*OrderInfo, len(orders))
  for i, order := range orders {
    data[i] = &OrderInfo{
      ID:              order.ID,
      Symbol:          order.Symbol,
      OrderId:         order.OrderId,
      Type:            order.Type,
      Side:            order.Side,
      Price:           order.Price,
      Quantity:        order.Quantity,
      OpenTime:        order.OpenTime,
      UpdateTime:      order.UpdateTime,
      Status:          order.Status,
      Timestamp:       order.CreatedAt.UnixMicro(),
      TimestampFormat: common.FormatDatetime(order.CreatedAt),
    }
  }

  h.Response.Pagenate(data, total, current, pageSize)
}

func (h *OrdersHandler) Create(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  symbol := r.URL.Query().Get("symbol")
  if symbol == "" {
    h.Response.Error(http.StatusForbidden, 1004, "symbol is empty")
    return
  }
  side := r.URL.Query().Get("side")
  if side == "" {
    h.Response.Error(http.StatusForbidden, 1004, "side is empty")
    return
  }
  if r.URL.Query().Get("price") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "price is empty")
    return
  }
  price, _ := strconv.ParseFloat(r.URL.Query().Get("price"), 64)
  if r.URL.Query().Get("amount") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "amount is empty")
    return
  }
  amount, _ := strconv.ParseFloat(r.URL.Query().Get("amount"), 64)

  _, err := h.Repository.Create(symbol, side, price, amount)
  if err != nil {
    h.Response.Error(http.StatusForbidden, 1004, err.Error())
    return
  }

  h.Response.Json(nil)
}

func (h *OrdersHandler) Cancel(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  //id := chi.URLParam(r, "id")
  //err := h.Repository.Cancel(id)
  //if err != nil {
  //	h.Response.Error(http.StatusForbidden, 1004, err.Error())
  //	return
  //}

  h.Response.Json(nil)
}
