package v1

import (
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/account/api"
  "taoniu.local/account/api/v1/mqtt"
  "taoniu.local/account/common"
)

func NewMqttRouter(apiContext *common.ApiContext) http.Handler {
  r := chi.NewRouter()
  r.Use(api.Authenticator)
  r.Mount("/auth", mqtt.NewAuthRouter(apiContext))
  r.Mount("/acl", mqtt.NewAclRouter(apiContext))
  return r
}
