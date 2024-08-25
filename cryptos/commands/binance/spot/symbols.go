package spot

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type SymbolsHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.SymbolsRepository
}

func NewSymbolsCommand() *cli.Command {
  var h SymbolsHandler
  return &cli.Command{
    Name:  "symbols",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SymbolsHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.SymbolsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "currencies",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Currencies(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
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
      {
        Name:  "adjust",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Adjust(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *SymbolsHandler) Currencies() error {
  log.Println("symbols currencies...")
  currencies := h.Repository.Currencies()
  log.Println("currencies", currencies)
  return nil
}

func (h *SymbolsHandler) Flush() error {
  log.Println("symbols flush...")
  return h.Repository.Flush()
}

func (h *SymbolsHandler) Price(symbol string) (err error) {
  log.Println("symbols price...")
  price, err := h.Repository.Price(symbol)
  if err != nil {
    return
  }
  log.Println("symbol price", symbol, price)
  return
}

func (h *SymbolsHandler) Count() error {
  log.Println("symbols count...")
  return h.Repository.Count()
}

func (h *SymbolsHandler) Slippage() error {
  log.Println("symbols depth...")
  for _, symbol := range h.Repository.Symbols() {
    h.Repository.Slippage(symbol)
  }
  return nil
}

func (h *SymbolsHandler) Adjust() error {
  log.Println("symbols adjust...")
  symbol := "AVAXUSDT"
  price := 11.81 * 1.02
  price, quantity, err := h.Repository.Adjust(symbol, price, 20)
  log.Println("price", price, quantity, err)
  return nil
}
