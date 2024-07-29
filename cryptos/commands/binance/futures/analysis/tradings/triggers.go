package tradings

import (
  "log"
  "strconv"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures/analysis/tradings"
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
        Db: common.NewDB(2),
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
          side, _ := strconv.Atoi(c.Args().Get(0))
          if side != 1 && side != 2 {
            log.Fatal("side invalid")
            return nil
          }
          if err := h.Flush(side); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *TriggersHandler) Flush(side int) error {
  log.Println("binance futures analysis tradings triggers flush...")
  err := h.Repository.Flush(side)
  if err != nil {
    log.Println("error", err)
  }
  return nil
}
