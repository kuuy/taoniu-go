package spot

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

type SymbolsHandler struct{}

func NewSymbolsRouter() http.Handler {
	r := chi.NewRouter()

	return r
}
