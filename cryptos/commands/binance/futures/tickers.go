package futures

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type TickersHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.TickersRepository
}

func NewTickersCommand() *cli.Command {
	var h TickersHandler
	return &cli.Command{
		Name:  "tickers",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = TickersHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.TickersRepository{
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

func (h *TickersHandler) flush() error {
	log.Println("Tickers flush...")
	symbols, _ := h.Rdb.ZRevRange(
		h.Ctx,
		"binance:futures:tickers:flush",
		0,
		-1,
	).Result()
	for i := 0; i < len(symbols); i += 20 {
		j := i + 20
		if j > len(symbols) {
			j = len(symbols)
		}
		h.Repository.Flush(symbols[i:j])
	}

	return nil
}
