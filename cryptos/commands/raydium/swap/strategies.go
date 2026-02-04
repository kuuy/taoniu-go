package swap

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	"taoniu.local/cryptos/repositories/raydium/swap"
)

type StrategiesHandler struct {
	Db                   *gorm.DB
	Rdb                  *redis.Client
	Ctx                  context.Context
	StrategiesRepository *swap.StrategiesRepository
}

func NewStrategiesCommand() *cli.Command {
	var h StrategiesHandler
	return &cli.Command{
		Name:  "strategies",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = StrategiesHandler{
				Db:  common.NewDB(3),
				Rdb: common.NewRedis(3),
				Ctx: context.Background(),
			}
			h.StrategiesRepository = &swap.StrategiesRepository{
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
					symbol := c.Args().Get(1)
					interval := c.Args().Get(0)
					if err := h.Flush(symbol, interval); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *StrategiesHandler) Flush(symbol string, interval string) (err error) {
	log.Println("raydium swap strategies flush...", symbol, interval)
	err = h.StrategiesRepository.Flush(symbol, interval)
	if err != nil {
		log.Println("strategies flush error", err)
	}
	return nil
}
