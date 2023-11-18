package v1

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/account/api"
  "taoniu.local/account/repositories"
)

type ProfileHandler struct {
  Response        *api.ResponseHandler
  Repository      repositories.ProfileRepository
  UsersRepository repositories.UsersRepository
}

type ProfileInfo struct {
  Nickname string `json:"nickname"`
  Email    string `json:"email"`
  Avatar   string `json:"avatar"`
}

func NewProfileRouter() http.Handler {
  h := ProfileHandler{}

  r := chi.NewRouter()
  r.Use(api.Authenticator)
  r.Get("/", h.Get)
  r.Put("/save", h.Save)
  r.Put("/save", h.Save)

  return r
}

func (h *ProfileHandler) Get(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  uid := r.Header.Get("uid")
  user, err := h.UsersRepository.Find(uid)
  if err != nil {
    h.Response.Error(http.StatusForbidden, 1000, "user not exists")
    return
  }

  profile, err := h.Repository.Find(uid)
  if err != nil {
    h.Response.Error(http.StatusForbidden, 1000, "profile not exists")
    return
  }

  h.Response.Json(&ProfileInfo{
    Nickname: profile.Nickname,
    Email:    user.Email,
    Avatar:   profile.Avatar,
  })
}

func (h *ProfileHandler) Save(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  uid := r.Header.Get("uid")
  user, err := h.UsersRepository.Find(uid)
  if err != nil {
    h.Response.Error(http.StatusForbidden, 1000, "user not exists")
    return
  }

  r.ParseForm()

  d := r.Form
  if d.Get("nickname") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "nickname is empty")
    return
  }
  if d.Get("avatar") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "avatar is empty")
    return
  }

  profile, err := h.Repository.Find(uid)
  if err != nil {
    h.Response.Error(http.StatusForbidden, 1000, "profile not exists")
    return
  }

  h.Response.Json(&ProfileInfo{
    Nickname: profile.Nickname,
    Email:    user.Email,
    Avatar:   profile.Avatar,
  })
}
