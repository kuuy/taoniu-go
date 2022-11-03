package account

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api"
)

type LogoutHandler struct{}

func NewLogoutRouter() http.Handler {
	h := LogoutHandler{}

	r := chi.NewRouter()
	r.Use(api.Authenticator)
	r.Get("/", h.Do)

	return r
}

func (h *LogoutHandler) Do(
	w http.ResponseWriter,
	r *http.Request,
) {

}
