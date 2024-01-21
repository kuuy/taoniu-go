package commands

import (
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"
  "taoniu.local/groceries/common"
  "taoniu.local/groceries/models"
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
        Db: common.NewDB(),
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
    &models.Store{},
    &models.Product{},
    &models.Barcode{},
    &models.Feed{},
  )
  return nil
}
