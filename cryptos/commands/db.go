package commands

import (
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
	"log"
	pool "taoniu.local/cryptos/common"
	models "taoniu.local/cryptos/models"
)

type DbHandler struct {
	db *gorm.DB
}

func NewDbCommand() *cli.Command {
	handler := DbHandler{
		db: pool.NewDB(),
	}

	return &cli.Command{
		Name:  "db",
		Usage: "",
		Subcommands: []*cli.Command{
			{
				Name:  "migrate",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := handler.migrate(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *DbHandler) migrate() error {
	log.Println("process migrator")
	models.NewBinance().AutoMigrate(h.db)
	return nil
}
