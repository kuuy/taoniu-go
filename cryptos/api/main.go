package main

import (
  "log"
  "net/http"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api/routers"
)

func main() {
  log.Println("start api service")

  r := chi.NewRouter()
  r.Mount("/orders", routers.NewOrderRouter())
  r.Mount("/strategies", routers.NewStrategyRouter())
  http.ListenAndServe(":3000", r)
}

