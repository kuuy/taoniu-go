package v1

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/account/api"
)

type ProfileHandler struct{}

func NewProfileRouter() http.Handler {
	h := ProfileHandler{}

	r := chi.NewRouter()
	r.Use(api.Authenticator)
	r.Get("/", h.Show)

	return r
}

func (h *ProfileHandler) Show(
	w http.ResponseWriter,
	r *http.Request,
) {

}
