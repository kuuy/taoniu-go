package account

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

type TokenHandler struct{}

func NewTokenRouter() http.Handler {
	h := LoginHandler{}

	r := chi.NewRouter()
	r.Get("/refresh", h.Refresh)

	return r
}

func (h *LoginHandler) Refresh(
	w http.ResponseWriter,
	r *http.Request,
) {

}
