package mqtt

import (
  "log"
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/account/api"
  "taoniu.local/account/common"
)

type AclHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
}

func NewAclRouter(apiContext *common.ApiContext) http.Handler {
  h := AclHandler{
    ApiContext: apiContext,
  }

  r := chi.NewRouter()
  r.Post("/", h.Do)

  return r
}

func (h *AclHandler) Do(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  log.Println("body", r.Header, r.URL.RawQuery)

  h.Response.Json(nil)
}
