package mqtt

import (
  "log"
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/account/api"
  "taoniu.local/account/common"
)

type AuthHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
}

func NewAuthRouter(apiContext *common.ApiContext) http.Handler {
  h := AuthHandler{
    ApiContext: apiContext,
  }

  r := chi.NewRouter()
  r.Post("/", h.Do)

  return r
}

func (h *AuthHandler) Do(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  log.Println("body", r.Header, r.URL.RawQuery)

  h.Response.Json(nil)
}
