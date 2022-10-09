package profits

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/analysis/margin/isolated/profits"
)

type DailyHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.DailyRepository
}

func NewDailyCommand() *cli.Command {
	h := DailyHandler{
		Rdb: pool.NewRedis(),
		Ctx: context.Background(),
		Repository: &repositories.DailyRepository{
			Db:  pool.NewDB(),
			Rdb: pool.NewRedis(),
			Ctx: context.Background(),
		},
	}

	return &cli.Command{
		Name:  "daily",
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

func (h *DailyHandler) flush() error {
	log.Println("analysis margin isolated daily...")
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:spot:margin:isolated:symbols").Result()
	for _, symbol := range symbols {
		h.Repository.Flush(symbol)
	}
	return nil
}
