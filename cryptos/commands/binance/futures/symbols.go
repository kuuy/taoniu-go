package futures

import (
  "context"
  "github.com/urfave/cli/v2"
  "log"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type SymbolsHandler struct {
  SymbolsRepository *repositories.SymbolsRepository
}

func NewSymbolsCommand() *cli.Command {
  var h SymbolsHandler
  return &cli.Command{
    Name:  "symbols",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SymbolsHandler{}
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
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
        Name:  "price",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if err := h.Price(symbol); err != nil {
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
  return h.SymbolsRepository.Flush()
}

func (h *SymbolsHandler) Price(symbol string) (err error) {
  log.Println("symbols price...")
  price, err := h.SymbolsRepository.Price(symbol)
  if err != nil {
    return
  }
  log.Println("symbol price", symbol, price)
  return
}

func (h *SymbolsHandler) Count() error {
  log.Println("symbols count...")
  return h.SymbolsRepository.Count()
}

func (h *SymbolsHandler) Slippage() error {
  log.Println("symbols depth...")
  for _, symbol := range h.SymbolsRepository.Symbols() {
    symbol = "XVGUSDT"
    err := h.SymbolsRepository.Slippage(symbol)
    if err != nil {
      log.Println("error", err.Error())
    }
    break
  }
  return nil
}
