package v1

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/gamblings/api/v1/wolf"
)

func NewWolfRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/dice", wolf.NewDiceRouter())

	return r
}
