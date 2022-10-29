package spot

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	"strconv"
	"strings"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
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

func (h *OrdersHandler) open() error {
	log.Println("spot open orders...")
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		log.Println("symbol:", symbol)
		h.Repository.Open(symbol)
	}
	return nil
}

func (h *OrdersHandler) flush() error {
	log.Println("margin orders flush...")
	orders, err := h.Rdb.SMembers(h.Ctx, "binance:spot:orders:flush").Result()
	if err != nil {
		return nil
	}
	for _, order := range orders {
		data := strings.Split(order, ",")
		symbol := data[0]
		orderID, _ := strconv.ParseInt(data[1], 10, 64)
		h.Repository.Flush(symbol, orderID)
	}

	return nil
}
