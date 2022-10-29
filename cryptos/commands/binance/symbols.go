package binance

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance"
)

type SymbolsHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.SymbolsRepository
}

func NewSymbolsCommand() *cli.Command {
	var h SymbolsHandler
	return &cli.Command{
		Name:  "symbols",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = SymbolsHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.SymbolsRepository{
				Db:  pool.NewDB(),
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			return nil
		},
		Subcommands: []*cli.Command{
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
			{
				Name:  "count",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.count(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *SymbolsHandler) flush() error {
	log.Println("symbols flush...")
	return h.Repository.Flush()
}

func (h *SymbolsHandler) count() error {
	log.Println("symbols count...")
	return h.Repository.Count()
}
