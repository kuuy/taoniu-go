package spot

import (
  "log"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type PlansHandler struct {
  Db                *gorm.DB
  Repository        *repositories.PlansRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewPlansCommand() *cli.Command {
  var h PlansHandler
  return &cli.Command{
    Name:  "plans",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = PlansHandler{
        Db: common.NewDB(1),
      }
      h.Repository = &repositories.PlansRepository{
        Db: h.Db,
      }
      h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          interval := c.Args().Get(0)
          if err := h.Flush(interval); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "clean",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Clean(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *PlansHandler) Flush(interval string) error {
  log.Println("spot plans flush...")
  return h.Repository.Flush(interval)
}

func (h *PlansHandler) Clean() error {
  log.Println("binance spot plans clean...")
  symbols := h.SymbolsRepository.Symbols()
  for _, symbol := range symbols {
    h.Repository.Clean(symbol)
  }
  return nil
}
