package commands

import (
	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
	"log"
	"net/http"
	"taoniu.local/cryptos/api"
	"taoniu.local/cryptos/api/v1"
	pool "taoniu.local/cryptos/common"
)

type ApiHandler struct {
	db *gorm.DB
}

func NewApiCommand() *cli.Command {
	h := ApiHandler{
		db: pool.NewDB(),
	}

	return &cli.Command{
		Name:  "api",
		Usage: "",
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
	r.Mount("/orders", api.NewOrderRouter())
	r.Mount("/strategies", api.NewStrategyRouter())

	r.Route("/v1", func(r chi.Router) {
		r.Mount("/strategies", v1.NewStrategyRouter())
	})

	http.ListenAndServe(":3000", r)

	return nil
}
