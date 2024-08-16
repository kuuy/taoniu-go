package v1

import (
  "encoding/json"
  "io"
  "net/http"

  "github.com/go-chi/chi/v5"
  "taoniu.local/account/api"
  "taoniu.local/account/common"
  "taoniu.local/account/repositories"
)

type LoginHandler struct {
  ApiContext      *common.ApiContext
  Response        *api.ResponseHandler
  JweRepository   *repositories.JweRepository
  UserRepository  *repositories.UsersRepository
  TokenRepository *repositories.TokenRepository
}

func NewLoginRouter(apiContext *common.ApiContext) http.Handler {
  h := LoginHandler{
    ApiContext: apiContext,
  }
  h.Response = &api.ResponseHandler{}
  h.Response.JweRepository = &repositories.JweRepository{}
  h.JweRepository = h.Response.JweRepository
  h.UserRepository = &repositories.UsersRepository{
    Db: h.ApiContext.Db,
  }

  r := chi.NewRouter()
  r.Post("/", h.Do)

  return r
}

func (h *LoginHandler) Token() *repositories.TokenRepository {
  if h.TokenRepository == nil {
    h.TokenRepository = &repositories.TokenRepository{}
  }
  return h.TokenRepository
}

func (h *LoginHandler) Do(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{}
  h.Response.JweRepository = h.JweRepository
  h.Response.Writer = w

  body, _ := io.ReadAll(r.Body)
  payload, err := h.JweRepository.Decrypt(string(body))
  if err != nil {
    h.Response.Error(http.StatusForbidden, 1004, "bad request")
    return
  }

  var request *LoginRequest
  json.Unmarshal(payload, &request)

  if request.Email == "" {
    h.Response.Error(http.StatusForbidden, 1004, "email is empty")
    return
  }

  if request.Password == "" {
    h.Response.Error(http.StatusForbidden, 1004, "password is empty")
    return
  }

  user := h.UserRepository.Get(request.Email)
  if user == nil {
    h.Response.Error(http.StatusForbidden, 1000, "email or password not exists")
    return
  }
  if !common.VerifyPassword(request.Password, user.Salt, user.Password) {
    h.Response.Error(http.StatusForbidden, 1000, "email or password not exists")
    return
  }

  accessToken, err := h.Token().AccessToken(user.ID)
  if err != nil {
    h.Response.Error(http.StatusInternalServerError, 500, "server error")
    return
  }
  refreshToken, err := h.Token().RefreshToken(user.ID)
  if err != nil {
    h.Response.Error(http.StatusInternalServerError, 500, "server error")
    return
  }

  response := &LoginResponse{
    AccessToken:  accessToken,
    RefreshToken: refreshToken,
  }

  h.Response.Json(response)
}
