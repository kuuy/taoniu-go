package futures

import (
  "context"
  "github.com/urfave/cli/v2"
  "log"
  pool "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type SymbolsHandler struct {
  Repository *repositories.SymbolsRepository
}

func NewSymbolsCommand() *cli.Command {
  var h SymbolsHandler
  return &cli.Command{
    Name:  "symbols",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SymbolsHandler{}
      h.Repository = &repositories.SymbolsRepository{
        Db:  pool.NewDB(),
        Rdb: pool.NewRedis(),
        Ctx: context.Background(),
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
      {
        Name:  "count",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Count(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "slippage",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Slippage(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *SymbolsHandler) Flush() error {
  log.Println("symbols flush...")
  return h.Repository.Flush()
}

func (h *SymbolsHandler) Count() error {
  log.Println("symbols count...")
  return h.Repository.Count()
}

func (h *SymbolsHandler) Slippage() error {
  log.Println("symbols depth...")
  for _, symbol := range h.Repository.Symbols() {
    symbol = "XVGUSDT"
    err := h.Repository.Slippage(symbol)
    if err != nil {
      log.Println("error", err.Error())
    }
    break
  }
  return nil
}
