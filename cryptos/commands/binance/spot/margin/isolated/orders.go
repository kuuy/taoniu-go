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
	var h OrdersHandler
	return &cli.Command{
		Name:  "orders",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = OrdersHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.OrdersRepository{
				Db:  pool.NewDB(),
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			return nil
		},
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
			{
				Name:  "sync",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.sync(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *OrdersHandler) open() error {
	log.Println("margin isolated open orders...")
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		log.Println("symbol:", symbol)
		h.Repository.Open(symbol)
	}
	return nil
}

func (h *OrdersHandler) sync() error {
	log.Println("margin isolated sync orders...")
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:spot:margin:isolated:symbols").Result()
	for _, symbol := range symbols {
		h.Repository.Sync(symbol, 100)
	}
	return nil
}
