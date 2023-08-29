package futures

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type DepthHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  Repository         *repositories.DepthRepository
  SymbolsRepository  *repositories.SymbolsRepository
  TradingsRepository *repositories.TradingsRepository
}

func NewDepthCommand() *cli.Command {
  var h DepthHandler
  return &cli.Command{
    Name:  "depth",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = DepthHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.DepthRepository{
        Db: h.Db,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      h.TradingsRepository = &repositories.TradingsRepository{
        Db: h.Db,
      }
      h.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
        Db: h.Db,
      }
      h.TradingsRepository.TriggersRepository = &tradingsRepositories.TriggersRepository{
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
          h.Repository.UseProxy = c.Bool("proxy")
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
    err := h.Repository.Flush(symbol, 1000)
    if err != nil {
      log.Println("error", err)
    }
  }
  return nil
}

func (h *DepthHandler) Scan() []string {
  var symbols []string
  for _, symbol := range h.TradingsRepository.Scan() {
    if !h.contains(symbols, symbol) {
      symbols = append(symbols, symbol)
    }
  }
  return symbols
}

func (h *DepthHandler) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
