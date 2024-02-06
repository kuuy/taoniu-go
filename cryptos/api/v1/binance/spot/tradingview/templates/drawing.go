package templates

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
)

type DrawingHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
}

func NewDrawingRouter(apiContext *common.ApiContext) http.Handler {
  h := DrawingHandler{
    ApiContext: apiContext,
  }

  r := chi.NewRouter()
  r.Get("/", h.Gets)

  return r
}

func (h *DrawingHandler) Gets(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  h.Response.Out("")
}
