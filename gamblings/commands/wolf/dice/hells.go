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

type HellsHandler struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.HellsRepository
}

func NewHellsCommand() *cli.Command {
	var h HellsHandler
	return &cli.Command{
		Name:  "hells",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = HellsHandler{
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.HellsRepository{
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

func (h *HellsHandler) apply(currency string) error {
	log.Println("wolf dice bet hells apply...")
	return h.Repository.Apply(currency)
}

func (h *HellsHandler) start() error {
	log.Println("wolf dice bet plan starting...")
	h.Repository.Start()
	return nil
}

func (h *HellsHandler) stop() error {
	log.Println("wolf dice bet plan stopping...")
	h.Repository.Stop()
	return nil
}

func (h *HellsHandler) place() error {
	log.Println("wolf dice hells place...")
	return h.Repository.Place()
}

func (h *HellsHandler) rescue() error {
	log.Println("wolf dice bet hells rescue...")
	return h.Repository.Rescue()
}

func (h *HellsHandler) test() error {
	log.Println("wolf dice bet hells test...")
	rule := "over"
	betCount := 6
	rule = h.Repository.Switch(rule, betCount)
	return nil
}
