package account

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api"
	repositories "taoniu.local/cryptos/repositories/account"
)

type TokenHandler struct {
	Response        *api.ResponseHandler
	TokenRepository *repositories.TokenRepository
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
}

func NewTokenRouter() http.Handler {
	h := TokenHandler{}

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
	h.Response = &api.ResponseHandler{
		Writer: w,
	}

	r.ParseMultipartForm(1024)

	if r.Form.Get("refresh_token") == "" {
		h.Response.Error(http.StatusForbidden, 1004, "refresh token is empty")
		return
	}

	refreshToken := r.Form.Get("refresh_token")
	uid, err := h.Token().Uid(refreshToken)
	if err != nil {
		h.Response.Error(http.StatusForbidden, 403, "refresh token not valid")
		return
	}

	accessToken, err := h.Token().AccessToken(uid)
	if err != nil {
		h.Response.Error(http.StatusInternalServerError, 500, "server error")
		return
	}

	token := &AccessToken{
		AccessToken: accessToken,
	}

	h.Response.Json(token)
}
