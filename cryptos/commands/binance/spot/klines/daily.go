package klines

import (
	"context"
	"gorm.io/gorm"
	"log"
	"strconv"
	pool "taoniu.local/cryptos/common"
	models "taoniu.local/cryptos/models/binance"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	repositories "taoniu.local/cryptos/repositories/binance/spot/klines"
)

type DailyHandler struct {
	Db         *gorm.DB
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
				Db:  pool.NewDB(),
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.DailyRepository{
				Db:  h.Db,
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
					limit, _ := strconv.Atoi(c.Args().Get(0))
					if limit < 1 || limit > 100 {
						log.Fatal("limit not in 1~100")
						return nil
					}
					if err := h.flush(limit); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "clean",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.clean(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *DailyHandler) flush(limit int) error {
	log.Println("binance spot klines daily flush...")
	var symbols []string
	h.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		err := h.Repository.Flush(symbol, limit)
		if err != nil {
			log.Println("kline flush error", err)
		}
	}

	return nil
}

func (h *DailyHandler) clean() error {
	log.Println("binance spot klines daily clean...")
	h.Repository.Clean()
	return nil
}
