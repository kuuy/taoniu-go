package spot

import (
	"context"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type AccountHandler struct {
	Repository *repositories.AccountRepository
}

func NewAccountCommand() *cli.Command {
	h := AccountHandler{
		Repository: &repositories.AccountRepository{
			Rdb: pool.NewRedis(),
			Ctx: context.Background(),
		},
	}

	return &cli.Command{
		Name:  "account",
		Usage: "",
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

func (h *AccountHandler) flush() error {
	log.Println("account flush processing...")
	return h.Repository.Flush()
}
