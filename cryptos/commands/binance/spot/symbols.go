package spot

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type SymbolsHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.SymbolsRepository
}

func NewSymbolsCommand() *cli.Command {
	h := SymbolsHandler{
		Rdb: pool.NewRedis(),
		Ctx: context.Background(),
		Repository: &repositories.SymbolsRepository{
			Rdb: pool.NewRedis(),
			Ctx: context.Background(),
		},
	}

	return &cli.Command{
		Name:  "symbols",
		Usage: "",
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
		},
	}
}

func (h *SymbolsHandler) flush() error {
	log.Println("symbols flush...")
	return h.Repository.Flush()
}
