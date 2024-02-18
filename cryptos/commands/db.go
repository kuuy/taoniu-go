package commands

import (
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/models"
)

type DbHandler struct {
  Db *gorm.DB
}

func NewDbCommand() *cli.Command {
  var h DbHandler
  return &cli.Command{
    Name:  "db",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = DbHandler{
        Db: common.NewDB(1),
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "migrate",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.migrate(); err != nil {
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
  h.Db.AutoMigrate(
    &models.Currency{},
    &models.Sector{},
    &models.Exchange{},
  )
  models.NewBinance().AutoMigrate(h.Db)
  models.NewDydx().AutoMigrate(h.Db)
  models.NewTradingView().AutoMigrate(h.Db)
  models.NewSpiders().AutoMigrate(h.Db)
  return nil
}
