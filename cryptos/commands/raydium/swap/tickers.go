package swap

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	"taoniu.local/cryptos/repositories/raydium/swap"
)

type TickersHandler struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	TickersRepository *swap.TickersRepository
}

func NewTickersCommand() *cli.Command {
	var h TickersHandler
	return &cli.Command{
		Name:  "tickers",
		Usage: "Raydium swap tickers management",
		Before: func(c *cli.Context) error {
			h = TickersHandler{
				Db:  common.NewDB(3),
				Rdb: common.NewRedis(3),
				Ctx: context.Background(),
			}
			h.TickersRepository = &swap.TickersRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "flush",
				Usage: "Flush tickers from Raydium API to Redis",
				Action: func(c *cli.Context) error {
					if err := h.TickersRepository.Flush(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}
