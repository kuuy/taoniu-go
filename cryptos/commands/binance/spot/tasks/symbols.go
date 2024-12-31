package tasks

import (
  "context"
  "log"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type SymbolsHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  Nats               *nats.Conn
  SymbolsRepository  *repositories.SymbolsRepository
  TradingsRepository *repositories.TradingsRepository
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
          if err := h.Flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *SymbolsHandler) Flush() (err error) {
  log.Println("binance spot tasks symbols flush...")
  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    config.LOCKS_TASKS_SYMBOLS_FLUSH,
  )
  if !mutex.Lock(50 * time.Second) {
    return
  }
  err = h.SymbolsRepository.Flush()
  return
}
