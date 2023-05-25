package tradings

import (
  "github.com/go-chi/chi/v5"
  "net/http"
)

func NewFishersRouter() http.Handler {
  r := chi.NewRouter()
  return r
}
