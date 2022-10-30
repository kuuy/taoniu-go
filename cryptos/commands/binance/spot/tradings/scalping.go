package tradings

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type ScalpingHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.ScalpingRepository
}

func NewScalpingCommand() *cli.Command {
	var h ScalpingHandler
	return &cli.Command{
		Name:  "scalping",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = ScalpingHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.ScalpingRepository{
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

func (h *ScalpingHandler) flush() error {
	log.Println("spot tradings scalping flush...")
	h.Repository.Flush()
	return nil
}
