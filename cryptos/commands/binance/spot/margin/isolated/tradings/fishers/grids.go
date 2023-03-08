package fishers

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	savingsRepositories "taoniu.local/cryptos/repositories/binance/savings"
	isolatedRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings/fishers"
)

type GridsHandler struct {
	Db         *gorm.DB
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
				Db:  common.NewDB(),
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.GridsRepository{
				Db: h.Db,
			}
			h.Repository.AccountRepository = &isolatedRepositories.AccountRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.Repository.ProductsRepository = &savingsRepositories.ProductsRepository{
				Db: h.Db,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "pending",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Pending(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "collect",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Collect(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *GridsHandler) Pending() error {
	log.Println("spot margin isolated tradings fishers grids pending...")
	data := h.Repository.Pending()
	log.Println(data)
	return nil
}

func (h *GridsHandler) Collect() error {
	log.Println("spot margin isolated tradings fishers grids collect...")
	return h.Repository.Collect()
}
