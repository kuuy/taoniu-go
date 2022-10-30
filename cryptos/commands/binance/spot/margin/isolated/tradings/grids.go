package tradings

import (
	"context"
	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
)

type GridsHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.GridsRepository
}

func NewGridsCommand() *cli.Command {
	var h GridsHandler
	return &cli.Command{
		Name:  "grids",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = GridsHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.GridsRepository{
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
				Name:  "buy",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.buy(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *GridsHandler) flush() error {
	log.Println("spot margin isolated tradings grids flush...")
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:spot:margin:isolated:symbols").Result()
	for _, symbol := range symbols {
		err := h.Repository.Flush(symbol)
		if err != nil {
			log.Println("error", err)
		}
	}

	return nil
}

func (h *GridsHandler) buy() error {
	symbol := "AVAXBUSD"
	price := 15.427547306193494
	orderId, err := h.Repository.Order(symbol, binance.SideTypeBuy, price, 10)
	if err != nil {
		return err
	}
	log.Println("order:", symbol, orderId)

	return nil
}
