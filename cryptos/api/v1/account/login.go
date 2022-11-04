package account

import (
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"net/http"
	"taoniu.local/cryptos/api"
	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/account"
)

type LoginHandler struct {
	Db              *gorm.DB
	Response        *api.ResponseHandler
	UserRepository  *repositories.UsersRepository
	TokenRepository *repositories.TokenRepository
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewLoginRouter() http.Handler {
	h := LoginHandler{
		Db: common.NewDB(),
	}

	r := chi.NewRouter()
	r.Post("/", h.Do)

	return r
}

func (h *LoginHandler) Users() *repositories.UsersRepository {
	if h.UserRepository == nil {
		h.UserRepository = &repositories.UsersRepository{
			Db: h.Db,
		}
	}
	return h.UserRepository
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

	user := h.Users().Get(email)
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
