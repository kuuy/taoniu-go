package swap

import (
	"context"
	"log"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	"taoniu.local/cryptos/repositories/raydium/swap"
)

type IndicatorsHandler struct {
	Db                   *gorm.DB
	Rdb                  *redis.Client
	Ctx                  context.Context
	IndicatorsRepository *swap.IndicatorsRepository
}

func NewIndicatorsCommand() *cli.Command {
	var h IndicatorsHandler
	return &cli.Command{
		Name:  "indicators",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = IndicatorsHandler{
				Db:  common.NewDB(3),
				Rdb: common.NewRedis(3),
				Ctx: context.Background(),
			}
			h.IndicatorsRepository = &swap.IndicatorsRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
				KlinesRepository: &swap.KlinesRepository{
					Db:  h.Db,
					Rdb: h.Rdb,
					Ctx: h.Ctx,
				},
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "flush",
				Usage: "",
				Action: func(c *cli.Context) error {
					symbol := c.Args().Get(2)
					interval := c.Args().Get(0)
					limit, _ := strconv.Atoi(c.Args().Get(1))
					if interval == "1m" && (limit < 1 || limit > 1440) {
						log.Fatal("limit not in 1~1440")
						return nil
					}
					if interval == "15m" && (limit < 1 || limit > 672) {
						log.Fatal("limit not in 1~672")
						return nil
					}
					if interval == "4h" && (limit < 1 || limit > 126) {
						log.Fatal("limit not in 1~126")
						return nil
					}
					if interval == "1d" && (limit < 1 || limit > 100) {
						log.Fatal("limit not in 1~100")
						return nil
					}
					if err := h.Flush(symbol, interval, limit); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *IndicatorsHandler) Flush(symbol string, interval string, limit int) (err error) {
	log.Println("raydium swap indicators flush...", symbol, interval, limit)
	err = h.IndicatorsRepository.Flush(symbol, interval, limit)
	if err != nil {
		log.Println("indicators flush error", err)
	}
	return nil
}
