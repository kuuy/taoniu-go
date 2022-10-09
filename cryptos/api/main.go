package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	v1 "taoniu.local/cryptos/api/routers/v1"

	"taoniu.local/cryptos/api/routers"
	_ "taoniu.local/cryptos/api/routers/v1"
)

func main() {
	log.Println("start api service")

	r := chi.NewRouter()
	r.Mount("/orders", routers.NewOrderRouter())
	r.Mount("/strategies", routers.NewStrategyRouter())

	r.Route("/v1", func(r chi.Router) {
		r.Mount("/strategies", v1.NewStrategyRouter())
	})

	http.ListenAndServe(":3000", r)
}
