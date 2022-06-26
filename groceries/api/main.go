package main

import (
  "log"
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/groceries/api/routers"
)

func main() {
  log.Println("start api service")

  r := chi.NewRouter()
  r.Mount("/products", routers.NewProductRouter())
  http.ListenAndServe(":3090", r)
}

