package wolf

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/gamblings/common"
	repositories "taoniu.local/gamblings/repositories/wolf"
)

type AccountHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.AccountRepository
}

func NewAccountCommand() *cli.Command {
	var h AccountHandler
	return &cli.Command{
		Name:  "account",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = AccountHandler{
				Rdb: pool.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.AccountRepository{
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "balance",
				Usage: "",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "proxy",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					currency := c.Args().Get(0)
					if currency == "" {
						return errors.New("currency is empty")
					}
					h.Repository.UseProxy = c.Bool("proxy")
					if err := h.balance(currency); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *AccountHandler) balance(currency string) error {
	log.Println("wolf account balance...")
	balance, err := h.Repository.Balance(currency)
	if err != nil {
		log.Println("wolf account balance error", err)
	}
	log.Println("balance", balance)
	return nil
}
