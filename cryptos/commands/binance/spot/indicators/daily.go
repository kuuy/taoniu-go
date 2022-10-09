package indicators

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/indicators"
)

type DailyHandler struct {
	rdb        *redis.Client
	ctx        context.Context
	repository *repositories.DailyRepository
}

func NewDailyCommand() *cli.Command {
	h := DailyHandler{
		rdb: pool.NewRedis(),
		ctx: context.Background(),
		repository: &repositories.DailyRepository{
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
				Name:  "pivot",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.pivot(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
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
	symbols, _ := h.rdb.SMembers(h.ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		h.repository.Atr(symbol, 14, 100)
	}
	return nil
}

func (h *DailyHandler) zlema() error {
	log.Println("daily zlema processing...")
	symbols, _ := h.rdb.SMembers(h.ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		h.repository.Zlema(symbol, 14, 100)
	}
	return nil
}

func (h *DailyHandler) haZlema() error {
	log.Println("daily ha_zlema processing...")
	symbols, _ := h.rdb.SMembers(h.ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		h.repository.HaZlema(symbol, 14, 100)
	}
	return nil
}

func (h *DailyHandler) kdj() error {
	log.Println("daily kdj indicator...")
	symbols, _ := h.rdb.SMembers(h.ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		h.repository.Kdj(symbol, 9, 3, 100)
	}
	return nil
}

func (h *DailyHandler) bBands() error {
	log.Println("daily boll bands indicator...")
	symbols, _ := h.rdb.SMembers(h.ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		h.repository.BBands(symbol, 14, 100)
	}
	return nil
}

func (h *DailyHandler) pivot() error {
	log.Println("daily pivot indicator...")
	symbols, _ := h.rdb.SMembers(h.ctx, "binance:spot:websocket:symbols").Result()
	for _, symbol := range symbols {
		h.repository.Pivot(symbol)
	}
	return nil
}
