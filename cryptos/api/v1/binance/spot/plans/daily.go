package plans

import (
  "github.com/go-chi/chi/v5"
  "net/http"

  "taoniu.local/cryptos/api"
)

type DailyHandler struct {
  Response *api.ResponseHandler
  //Repository *repositories.DailyRepository
}

type DailyInfo struct {
  ID              string  `json:"id"`
  Symbol          string  `json:"symbol"`
  Side            int     `json:"side"`
  Price           float64 `json:"price"`
  Quantity        float64 `json:"quantity"`
  Amount          float64 `json:"amount"`
  Status          int     `json:"status"`
  Timestamp       int64   `json:"timestamp"`
  TimestampFormat string  `json:"timestamp_fmt"`
}

func NewDailyRouter() http.Handler {
  h := DailyHandler{}
  //h.Repository = &repositories.DailyRepository{
  //	Db:  common.NewDB(),
  //	Rdb: common.NewRedis(),
  //	Ctx: context.Background(),
  //}

  r := chi.NewRouter()
  r.Get("/", h.Listings)

  return r
}

func (h *DailyHandler) Listings(
  w http.ResponseWriter,
  r *http.Request,
) {
  //h.Response = &api.ResponseHandler{
  //  Writer: w,
  //}
  //
  //var current int
  //if !r.URL.Query().Has("current") {
  //  current = 1
  //}
  //current, _ = strconv.Atoi(r.URL.Query().Get("current"))
  //if current < 1 {
  //  h.Response.Error(http.StatusForbidden, 1004, "current not valid")
  //  return
  //}
  //
  //var pageSize int
  //if !r.URL.Query().Has("page_size") {
  //  pageSize = 50
  //} else {
  //  pageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
  //}
  //if pageSize < 1 || pageSize > 100 {
  //  h.Response.Error(http.StatusForbidden, 1004, "page size not valid")
  //  return
  //}
  //
  //conditions := make(map[string]interface{})
  //
  //if r.URL.Query().Get("symbols") != "" {
  //  conditions["symbols"] = strings.Split(r.URL.Query().Get("symbols"), ",")
  //}
  //
  //total := h.Repository.Count(conditions)
  //plans := h.Repository.Listings(conditions, current, pageSize)
  //data := make([]*DailyInfo, len(plans))
  //for i, plan := range plans {
  //  data[i] = &DailyInfo{
  //    ID:              plan.ID,
  //    Symbol:          plan.Symbol,
  //    Side:            plan.Side,
  //    Price:           plan.Price,
  //    Quantity:        plan.Quantity,
  //    Amount:          plan.Amount,
  //    Status:          plan.Status,
  //    Timestamp:       plan.CreatedAt.Unix(),
  //    TimestampFormat: common.FormatDatetime(plan.CreatedAt),
  //  }
  //}
  //
  //h.Response.Pagenate(data, total, current, pageSize)
}
