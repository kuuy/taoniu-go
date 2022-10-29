package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TradingsHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.TradingsRepository
}

func NewTradingsCommand() *cli.Command {
	var h TradingsHandler
	return &cli.Command{
		Name:  "tradings",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = TradingsHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.TradingsRepository{
				Db:  pool.NewDB(),
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "scalping",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.scalping(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *TradingsHandler) scalping() error {
	log.Println("spot tradings scalping...")
	h.Repository.Scalping()
	return nil
}
