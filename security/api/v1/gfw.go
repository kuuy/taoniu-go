package v1

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/security/api/v1/gfw"
)

func NewGfwRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/dns", gfw.NewDnsRouter())

	return r
}
