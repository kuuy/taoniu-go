package strategies

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/futures/strategies"
)

type DailyHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.DailyRepository
}

func NewDailyCommand() *cli.Command {
	var h DailyHandler
	return &cli.Command{
		Name:  "daily",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = DailyHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.DailyRepository{
				Db:  pool.NewDB(),
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "atr",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.atr(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "zlema",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.zlema(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "ha_zlema",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.haZlema(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "kdj",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.kdj(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "bbands",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.bBands(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *DailyHandler) atr() error {
	log.Println("daily atr processing...")
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:futures:websocket:symbols").Result()
	for _, symbol := range symbols {
		h.Repository.Atr(symbol)
	}
	return nil
}

func (h *DailyHandler) zlema() error {
	log.Println("daily zlema processing...")
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:futures:websocket:symbols").Result()
	for _, symbol := range symbols {
		h.Repository.Zlema(symbol)
	}
	return nil
}

func (h *DailyHandler) haZlema() error {
	log.Println("daily haZlema strategy...")
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:futures:websocket:symbols").Result()
	for _, symbol := range symbols {
		h.Repository.HaZlema(symbol)
	}
	return nil
}

func (h *DailyHandler) kdj() error {
	log.Println("daily zlema strategy...")
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:futures:websocket:symbols").Result()
	for _, symbol := range symbols {
		h.Repository.Kdj(symbol)
	}
	return nil
}

func (h *DailyHandler) bBands() error {
	log.Println("daily bbands strategy...")
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:futures:websocket:symbols").Result()
	for _, symbol := range symbols {
		h.Repository.BBands(symbol)
	}
	return nil
}
