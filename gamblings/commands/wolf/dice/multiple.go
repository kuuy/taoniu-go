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
				Name:  "place",
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
					if err := h.place(currency); err != nil {
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
				Name:  "monitor",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.monitor(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *MultipleHandler) place(currency string) error {
	log.Println("wolf dice multiple place...")

	for {
		score, _ := h.Rdb.ZScore(
			h.Ctx,
			"wolf:multiple",
			"dice",
		).Result()
		if int64(score) == 0 {
			return errors.New("bet multiple not started")
		}

		err := h.Repository.Place(currency)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *MultipleHandler) start() error {
	log.Println("wolf dice bet multiple starting...")
	return h.Repository.Start()
}

func (h *MultipleHandler) stop() error {
	log.Println("wolf dice bet multiple stopping...")
	return h.Repository.Stop()
}

func (h *MultipleHandler) monitor() error {
	log.Println("wolf dice bet multiple monitor...")

	return nil
}
