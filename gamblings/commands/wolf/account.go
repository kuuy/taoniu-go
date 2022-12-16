package wolf

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/gamblings/common"
	config "taoniu.local/gamblings/config/wolf"
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
			{
				Name:  "auth",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.auth(); err != nil {
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

func (h *AccountHandler) auth() error {
	log.Println("wolf account auth...")
	h.Rdb.HMSet(h.Ctx, "wolf:auth", map[string]string{
		"login_token":  config.LOGIN_TOKEN,
		"login_hash":   config.LOGIN_HASH,
		"login_cookie": config.LOGIN_COOKIE,
	})
	auth, _ := h.Rdb.HGetAll(h.Ctx, "wolf:auth").Result()
	log.Println("auth", auth["login_token"], auth["login_hash"], auth["login_cookie"])
	return nil
}
