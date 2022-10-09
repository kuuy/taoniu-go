package isolated

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type OrdersHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.OrdersRepository
}

func NewOrdersCommand() *cli.Command {
	h := OrdersHandler{
		Rdb: pool.NewRedis(),
		Ctx: context.Background(),
		Repository: &repositories.OrdersRepository{
			Db:  pool.NewDB(),
			Rdb: pool.NewRedis(),
			Ctx: context.Background(),
		},
	}

	return &cli.Command{
		Name:  "orders",
		Usage: "",
		Subcommands: []*cli.Command{
			{
				Name:  "open",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.open(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *OrdersHandler) open() error {
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		log.Println("symbol:", symbol)
		h.Repository.Open(symbol)
	}
	return nil
}
