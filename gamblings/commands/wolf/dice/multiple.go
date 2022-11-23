package dice

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"log"
	"taoniu.local/gamblings/common"
	repositories "taoniu.local/gamblings/repositories/wolf/dice"
)

type MultipleHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.MultipleRepository
}

func NewMultipleCommand() *cli.Command {
	var h MultipleHandler
	return &cli.Command{
		Name:  "multiple",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = MultipleHandler{
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.MultipleRepository{
				Db:  common.NewDB(),
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "apply",
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
					if err := h.apply(currency); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "place",
				Usage: "",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "proxy",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					h.Repository.UseProxy = c.Bool("proxy")
					if err := h.place(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "rescue",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.rescue(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "start",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.start(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.stop(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "test",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.test(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *MultipleHandler) apply(currency string) error {
	log.Println("wolf dice bet multiple apply...")
	return h.Repository.Apply(currency)
}

func (h *MultipleHandler) place() error {
	log.Println("wolf dice multiple place...")
	return h.Repository.Place()
}

func (h *MultipleHandler) rescue() error {
	log.Println("wolf dice bet multiple rescue...")
	return h.Repository.Rescue()
}

func (h *MultipleHandler) start() error {
	log.Println("wolf dice bet multiple starting...")
	h.Repository.Start()
	return nil
}

func (h *MultipleHandler) stop() error {
	log.Println("wolf dice bet multiple stopping...")
	h.Repository.Stop()
	return nil
}

func (h *MultipleHandler) test() error {
	log.Println("wolf dice bet multiple test...")
	rule := h.Repository.Switch(10)
	log.Println("rule", rule)
	return nil
}
