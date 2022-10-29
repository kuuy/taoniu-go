package savings

import (
	"context"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/savings"
)

type ProductsHandler struct {
	Repository *repositories.ProductsRepository
}

func NewProductsCommand() *cli.Command {
	var h ProductsHandler
	return &cli.Command{
		Name:  "products",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = ProductsHandler{}
			h.Repository = &repositories.ProductsRepository{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
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
		},
	}
}

func (h *ProductsHandler) flush() error {
	log.Println("savings products flush...")
	return h.Repository.Flush()
}
