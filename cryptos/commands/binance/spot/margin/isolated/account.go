package isolated

import (
	"context"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type AccountHandler struct {
	Repository *repositories.AccountRepository
}

func NewAccountCommand() *cli.Command {
	var h AccountHandler
	return &cli.Command{
		Name:  "account",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = AccountHandler{}
			h.Repository = &repositories.AccountRepository{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "flush",
				Usage: "",
				Before: func(c *cli.Context) error {
					return nil
				},
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

func (h *AccountHandler) flush() error {
	log.Println("margin isolated account flush processing...")
	return h.Repository.Flush()
}
