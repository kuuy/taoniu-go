package commands

import (
	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	v1 "taoniu.local/cryptos/api/v1"
)

type ApiHandler struct{}

func NewApiCommand() *cli.Command {
	var h ApiHandler
	return &cli.Command{
		Name:  "api",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = ApiHandler{}
			return nil
		},
		Action: func(c *cli.Context) error {
			if err := h.run(); err != nil {
				return cli.Exit(err.Error(), 1)
			}
			return nil
		},
	}
}

func (h *ApiHandler) run() error {
	log.Println("api running...")

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Mount("/binance", v1.NewBinanceRouter())
		r.Mount("/account", v1.NewAccountRouter())
	})

	http.ListenAndServe("127.0.0.1:3000", r)

	return nil
}
