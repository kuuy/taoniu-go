package dice

import (
	"context"
	"errors"
	"log"

	"github.com/urfave/cli/v2"

	"taoniu.local/gamblings/common"
	repositories "taoniu.local/gamblings/repositories/wolf/dice"
)

type BetHandler struct {
	Repository *repositories.BetRepository
}

func NewBetCommand() *cli.Command {
	var h BetHandler
	return &cli.Command{
		Name:  "bet",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = BetHandler{}
			h.Repository = &repositories.BetRepository{
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "",
				Action: func(c *cli.Context) error {
					strategy := c.Args().Get(0)
					if strategy == "" {
						return errors.New("strategy is empty")
					}
					if err := h.start(strategy); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "",
				Action: func(c *cli.Context) error {
					strategy := c.Args().Get(0)
					if strategy == "" {
						return errors.New("strategy is empty")
					}
					if err := h.stop(strategy); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *BetHandler) start(strategy string) error {
	log.Println("wolf dice bet starting...")
	h.Repository.Start(strategy)
	return nil
}

func (h *BetHandler) stop(strategy string) error {
	log.Println("wolf dice bet stopping...")
	h.Repository.Stop(strategy)
	return nil
}
