package v1

import (
  "net/http"

  "github.com/go-chi/chi/v5"
  "taoniu.local/account/api"
  "taoniu.local/account/common"
  "taoniu.local/account/repositories"
)

type LoginHandler struct {
  ApiContext      *common.ApiContext
  Response        *api.ResponseHandler
  UserRepository  *repositories.UsersRepository
  TokenRepository *repositories.TokenRepository
}

type Token struct {
  AccessToken  string `json:"access_token"`
  RefreshToken string `json:"refresh_token"`
}

func NewLoginRouter(apiContext *common.ApiContext) http.Handler {
  h := LoginHandler{
    ApiContext: apiContext,
  }
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
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  r.ParseMultipartForm(1024)

  if r.Form.Get("email") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "email is empty")
    return
  }

  if r.Form.Get("password") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "password is empty")
    return
  }

  email := r.Form.Get("email")
  password := r.Form.Get("password")

  user := h.UserRepository.Get(email)
  if user == nil {
    h.Response.Error(http.StatusForbidden, 1000, "email or password not exists")
    return
  }
  if !common.VerifyPassword(password, user.Salt, user.Password) {
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

  token := &Token{
    AccessToken:  accessToken,
    RefreshToken: refreshToken,
  }

  h.Response.Json(token)
}
