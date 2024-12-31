package futures

import (
  "context"
  "log"
  "slices"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type DepthHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  DepthRepository    *repositories.DepthRepository
  SymbolsRepository  *repositories.SymbolsRepository
  ScalpingRepository *repositories.ScalpingRepository
}

func NewDepthCommand() *cli.Command {
  var h DepthHandler
  return &cli.Command{
    Name:  "depth",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = DepthHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.DepthRepository = &repositories.DepthRepository{
        Db: h.Db,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      h.ScalpingRepository = &repositories.ScalpingRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Flags: []cli.Flag{
          &cli.BoolFlag{
            Name:  "proxy",
            Value: false,
          },
        },
        Action: func(c *cli.Context) error {
          h.DepthRepository.UseProxy = c.Bool("proxy")
          if err := h.Flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *DepthHandler) Flush() error {
  log.Println("symbols depth flush...")
  symbols := h.Scan()
  for _, symbol := range symbols {
    err := h.DepthRepository.Flush(symbol, 1000)
    if err != nil {
      log.Println("error", err)
    }
  }
  return nil
}

func (h *DepthHandler) Scan() []string {
  var symbols []string
  for _, symbol := range h.ScalpingRepository.Scan(2) {
    if !slices.Contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}
