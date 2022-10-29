package margin

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/analysis/daily/margin"
)

type IsolatedHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.IsolatedRepository
}

func NewIsolatedCommand() *cli.Command {
	var h IsolatedHandler
	return &cli.Command{
		Name:  "isolated",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = IsolatedHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.IsolatedRepository{
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
		},
	}
}

func (h *IsolatedHandler) flush() error {
	log.Println("analysis daily margin Isolated flush...")
	h.Repository.Grids()

	return nil
}
