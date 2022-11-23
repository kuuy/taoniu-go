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

type PlansHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.PlansRepository
}

func NewPlansCommand() *cli.Command {
	var h PlansHandler
	return &cli.Command{
		Name:  "plans",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = PlansHandler{
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.PlansRepository{
				Db:  common.NewDB(),
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			return nil
		},
		Subcommands: []*cli.Command{
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

func (h *PlansHandler) apply(currency string) error {
	log.Println("wolf dice bet plan apply...")
	return h.Repository.Apply(currency)
}

func (h *PlansHandler) place() error {
	log.Println("wolf dice plan place...")
	return h.Repository.Place()
}

func (h *PlansHandler) rescue() error {
	log.Println("wolf dice bet plans rescue...")
	return h.Repository.Rescue()
}

func (h *PlansHandler) start() error {
	log.Println("wolf dice bet plan starting...")
	h.Repository.Start()
	return nil
}

func (h *PlansHandler) stop() error {
	log.Println("wolf dice bet plan stopping...")
	h.Repository.Stop()
	return nil
}

func (h *PlansHandler) test() error {
	log.Println("wolf dice bet plan test...")
	rule := "over"
	betCount := 6
	rule = h.Repository.Switch(rule, betCount)
	return nil
}
