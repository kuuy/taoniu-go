package futures

import (
  "log"
  "time"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
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
        Db: common.NewDB(),
      }
      h.Repository = &repositories.TriggersRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "apply",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.apply(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *TriggersHandler) apply() error {
  log.Println("futures triggers apply...")
  symbol := "TOMOUSDT"
  side := 2
  capital := 5000.0
  price := 1.6350
  expiredAt := time.Now().Add(time.Hour * 24 * 14)
  err := h.Repository.Apply(symbol, side, capital, price, expiredAt)
  if err != nil {
    return err
  }

  return nil
}
