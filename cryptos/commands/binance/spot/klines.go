package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type KlinesHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.KlinesRepository
}

func NewKlinesCommand() *cli.Command {
	h := KlinesHandler{
		Rdb: pool.NewRedis(),
		Ctx: context.Background(),
		Repository: &repositories.KlinesRepository{
			Db:  pool.NewDB(),
			Rdb: pool.NewRedis(),
			Ctx: context.Background(),
		},
	}

	return &cli.Command{
		Name:  "klines",
		Usage: "",
		Subcommands: []*cli.Command{
			{
				Name:  "daily",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.daily(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *KlinesHandler) daily() error {
	log.Println("klines daily processing...")
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		h.Repository.FlushDaily(symbol, 100)
	}

	return nil
}
