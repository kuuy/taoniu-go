package spot

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
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
			h = SymbolsHandler{}
			h.Repository = &repositories.SymbolsRepository{
				Db:  common.NewDB(),
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "flush",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Flush(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "scan",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Scan(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "count",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Count(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *SymbolsHandler) Flush() error {
	log.Println("symbols flush...")
	return h.Repository.Flush()
}

func (h *SymbolsHandler) Scan() error {
	log.Println("symbols scan...")
	symbols := h.Repository.Scan()
	log.Println("symbols", symbols)
	return nil
}

func (h *SymbolsHandler) Count() error {
	log.Println("symbols count...")
	return h.Repository.Count()
}
