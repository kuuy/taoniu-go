package tradingview

import (
  "net/http"

  "github.com/go-chi/chi/v5"
  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
)

type ChartsHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
}

type ChartInfo struct {
  ID         string `json:"id"`
  Symbol     string `json:"symbol"`
  Name       string `json:"name"`
  Resolution string `json:"resolution"`
  Timestamp  int64  `json:"timestamp"`
}

func NewChartsRouter(apiContext *common.ApiContext) http.Handler {
  h := ChartsHandler{
    ApiContext: apiContext,
  }

  r := chi.NewRouter()
  r.Get("/", h.Gets)

  return r
}

func (h *ChartsHandler) Gets(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  var charts []*ChartInfo

  h.Response.Out(map[string]interface{}{
    "status": "ok",
    "data":   charts,
  })
}
