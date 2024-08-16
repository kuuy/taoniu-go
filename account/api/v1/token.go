package v1

import (
  "encoding/json"
  "io"
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/account/api"
  "taoniu.local/account/repositories"
)

type TokenHandler struct {
  Response        *api.ResponseHandler
  JweRepository   *repositories.JweRepository
  TokenRepository *repositories.TokenRepository
}

func NewTokenRouter() http.Handler {
  h := TokenHandler{}
  h.Response = &api.ResponseHandler{}
  h.Response.JweRepository = &repositories.JweRepository{}
  h.JweRepository = h.Response.JweRepository

  r := chi.NewRouter()
  r.Post("/refresh", h.Refresh)

  return r
}

func (h *TokenHandler) Token() *repositories.TokenRepository {
  if h.TokenRepository == nil {
    h.TokenRepository = &repositories.TokenRepository{}
  }
  return h.TokenRepository
}

func (h *TokenHandler) Refresh(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response.Writer = w

  body, _ := io.ReadAll(r.Body)
  payload, err := h.JweRepository.Decrypt(string(body))
  if err != nil {
    h.Response.Error(http.StatusForbidden, 1004, "bad request")
    return
  }
  var request *RefreshTokenRequest
  json.Unmarshal(payload, &request)

  if request.RefreshToken == "" {
    h.Response.Error(http.StatusForbidden, 1004, "token is empty")
    return
  }

  uid, err := h.Token().Uid(request.RefreshToken)
  if err != nil {
    if uid != "" {
      h.Response.Error(http.StatusForbidden, 401, err.Error())
    } else {
      h.Response.Error(http.StatusForbidden, 403, "token not valid")
    }
    return
  }

  accessToken, err := h.Token().AccessToken(uid)
  if err != nil {
    h.Response.Error(http.StatusInternalServerError, 500, "server error")
    return
  }

  response := &RefreshTokenResponse{
    AccessToken: accessToken,
  }

  h.Response.Json(response)
}
