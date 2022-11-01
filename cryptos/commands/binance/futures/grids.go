package futures

import (
	"context"
	"log"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type GridsHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.GridsRepository
}

func NewGridsCommand() *cli.Command {
	var h GridsHandler
	return &cli.Command{
		Name:  "grids",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = GridsHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.GridsRepository{
				Db:  pool.NewDB(),
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "open",
				Usage: "",
				Action: func(c *cli.Context) error {
					symbol := c.Args().Get(0)
					amount, _ := strconv.ParseFloat(c.Args().Get(1), 16)
					if symbol == "" {
						log.Fatal("grid symbol can not be empty")
						return nil
					}
					if amount < 50 {
						log.Fatal("grid amount less than 50")
						return nil
					}
					if err := h.open(symbol, amount); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "close",
				Usage: "",
				Action: func(c *cli.Context) error {
					symbol := c.Args().Get(0)
					if symbol == "" {
						log.Fatal("grid symbol can not be empty")
						return nil
					}
					if err := h.close(symbol); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "flush",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.flush(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *GridsHandler) open(symbol string, amount float64) error {
	log.Println("futures grids open...")
	return h.Repository.Open(symbol, amount)
}

func (h *GridsHandler) close(symbol string) error {
	log.Println("futures grids close...")
	return h.Repository.Close(symbol)
}

func (h *GridsHandler) flush() error {
	log.Println("futures grids flush...")
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:futures:grids:symbols").Result()
	for _, symbol := range symbols {
		h.Repository.Flush(symbol)
	}
	return nil
}
