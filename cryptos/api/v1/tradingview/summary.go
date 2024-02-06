package tradingview

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/tradingview"
)

type SummaryHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *repositories.AnalysisRepository
}

func NewSummaryRouter(apiContext *common.ApiContext) http.Handler {
  h := SummaryHandler{
    ApiContext: apiContext,
  }
  h.Repository = &repositories.AnalysisRepository{
    Db: h.ApiContext.Db,
  }

  r := chi.NewRouter()
  r.Get("/", h.Get)

  return r
}

func (h *SummaryHandler) Get(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  exchange := chi.URLParam(r, "exchange")
  symbol := chi.URLParam(r, "symbol")
  interval := chi.URLParam(r, "interval")
  summary, err := h.Repository.Summary(exchange, symbol, interval)
  if err != nil {
    http.Error(w, http.StatusText(404), http.StatusNotFound)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  h.Response.Json(summary)
}
