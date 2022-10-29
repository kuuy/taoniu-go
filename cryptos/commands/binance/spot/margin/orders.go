package margin

import (
	"context"
	"log"
	"strconv"
	"strings"
	pool "taoniu.local/cryptos/common"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	repositories "taoniu.local/cryptos/repositories/binance/spot/margin"
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
				Name:  "flush",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.flush(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "fix",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.fix(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *OrdersHandler) flush() error {
	log.Println("margin orders flush...")
	orders, err := h.Rdb.SMembers(h.Ctx, "binance:spot:margin:orders:flush").Result()
	if err != nil {
		return nil
	}
	for _, order := range orders {
		data := strings.Split(order, ",")
		symbol := data[0]
		orderID, _ := strconv.ParseInt(data[1], 10, 64)
		isIsolated, _ := strconv.ParseBool(data[2])
		h.Repository.Flush(symbol, orderID, isIsolated)
	}
	return nil
}

func (h *OrdersHandler) fix() error {
	log.Println("margin orders fix...")
	return h.Repository.Fix(time.Now(), 20)
}
