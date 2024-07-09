package tradings

import (
  "log"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/margin/cross/analysis/tradings"
)

type TriggersHandler struct {
  Db         *gorm.DB
  Repository *repositories.TriggersRepository
}

func NewTriggersCommand() *cli.Command {
  var h TriggersHandler
  return &cli.Command{
    Name:  "triggers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TriggersHandler{
        Db: common.NewDB(1),
      }
      h.Repository = &repositories.TriggersRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *TriggersHandler) Flush() error {
  log.Println("analysis tradings triggers flush...")
  err := h.Repository.Flush()
  if err != nil {
    log.Println("error", err)
  }
  return nil
}
