package plans

import (
	"context"
	"github.com/urfave/cli/v2"
	"log"
	pool "taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot/plans"
)

type DailyHandler struct {
	Repository *repositories.DailyRepository
}

func NewDailyCommand() *cli.Command {
	h := DailyHandler{
		Repository: &repositories.DailyRepository{
			Db:  pool.NewDB(),
			Rdb: pool.NewRedis(),
			Ctx: context.Background(),
		},
	}

	return &cli.Command{
		Name:  "daily",
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
			{
				Name:  "fix",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.fix(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *DailyHandler) flush() error {
	log.Println("spot plans daily flush...")
	return h.Repository.Flush()
}

func (h *DailyHandler) fix() error {
	log.Println("spot plans daily fix...")
	return h.Repository.Fix(7 * 86400)
}
