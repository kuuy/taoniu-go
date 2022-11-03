package v1

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/cryptos/api/v1/account"
)

func NewAccountRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/login", account.NewLoginRouter())
	r.Mount("/logout", account.NewLogoutRouter())
	r.Mount("/profile", account.NewProfileRouter())

	return r
}
