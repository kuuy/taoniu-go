package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	"strconv"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type GridsHandler struct {
	Symbol     string
	Amount     float64
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.GridsRepository
}

func NewGridsCommand() *cli.Command {
	h := GridsHandler{
		Rdb: pool.NewRedis(),
		Ctx: context.Background(),
		Repository: &repositories.GridsRepository{
			Db:  pool.NewDB(),
			Rdb: pool.NewRedis(),
			Ctx: context.Background(),
		},
	}

	return &cli.Command{
		Name:  "grids",
		Usage: "",
		Subcommands: []*cli.Command{
			{
				Name:  "open",
				Usage: "",
				Action: func(c *cli.Context) error {
					h.Symbol = c.Args().Get(0)
					h.Amount, _ = strconv.ParseFloat(c.Args().Get(1), 16)
					if h.Symbol == "" {
						log.Fatal("grid symbol can not be empty")
						return nil
					}
					if h.Amount < 50 {
						log.Fatal("grid amount less than 50")
						return nil
					}
					if err := h.open(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "close",
				Usage: "",
				Action: func(c *cli.Context) error {
					h.Symbol = c.Args().Get(0)
					if h.Symbol == "" {
						log.Fatal("grid symbol can not be empty")
						return nil
					}
					if err := h.close(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "flush",
				Usage: "",
				Action: func(c *cli.Context) error {
					h.Symbol = c.Args().Get(0)
					if h.Symbol == "" {
						log.Fatal("grid symbol can not be empty")
						return nil
					}
					if err := h.flush(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *GridsHandler) open() error {
	log.Println("spot grids open...")
	return h.Repository.Open(h.Symbol, h.Amount)
}

func (h *GridsHandler) close() error {
	log.Println("spot grids close...")
	return h.Repository.Close(h.Symbol)
}

func (h *GridsHandler) flush() error {
	log.Println("spot grids flush...")
	return h.Repository.Close(h.Symbol)
}
