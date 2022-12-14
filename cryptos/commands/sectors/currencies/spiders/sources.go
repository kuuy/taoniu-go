package spiders

import (
	"log"

	"gorm.io/gorm"

	"github.com/urfave/cli/v2"

	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/sectors/currencies/spiders"
)

type SourcesHandler struct {
	Db         *gorm.DB
	Repository *repositories.SourcesRepository
}

func NewSourcesCommand() *cli.Command {
	var h SourcesHandler
	return &cli.Command{
		Name:  "sources",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = SourcesHandler{
				Db: common.NewDB(),
			}
			h.Repository = &repositories.SourcesRepository{
				Db: h.Db,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.add(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *SourcesHandler) add() error {
	log.Println("sources add processing...")
	return h.Repository.Add()
}
