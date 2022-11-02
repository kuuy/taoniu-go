package futures

import (
	"context"
	"gorm.io/gorm"
	"log"
	"strconv"
	pool "taoniu.local/cryptos/common"
	models "taoniu.local/cryptos/models/binance/futures"
	repositories "taoniu.local/cryptos/repositories/binance/futures"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
)

type KlinesHandler struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.KlinesRepository
}

func NewKlinesCommand() *cli.Command {
	var h KlinesHandler
	return &cli.Command{
		Name:  "klines",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = KlinesHandler{
				Db:  pool.NewDB(),
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.KlinesRepository{
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
					interval := c.Args().Get(0)
					limit, _ := strconv.Atoi(c.Args().Get(1))
					if limit < 1 || limit > 100 {
						log.Fatal("limit not in 1~100")
						return nil
					}
					if err := h.flush(interval, limit); err != nil {
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

func (h *KlinesHandler) flush(interval string, limit int) error {
	log.Println("binance futures klines flush...")
	var symbols []string
	h.Db.Model(models.Symbol{}).Select("symbol").Where("status", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		err := h.Repository.Flush(symbol, interval, limit)
		if err != nil {
			log.Println("kline flush error", err)
		}
	}

	return nil
}

func (h *KlinesHandler) clean() error {
	log.Println("binance futures klines daily clean...")
	h.Repository.Clean()
	return nil
}
