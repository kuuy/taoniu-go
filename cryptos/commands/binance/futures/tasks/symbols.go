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
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type SymbolsHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  Nats               *nats.Conn
  Repository         *repositories.SymbolsRepository
  TradingsRepository *repositories.TradingsRepository
}

func NewSymbolsCommand() *cli.Command {
  var h SymbolsHandler
  return &cli.Command{
    Name:  "symbols",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SymbolsHandler{
        Db:   common.NewDB(2),
        Rdb:  common.NewRedis(2),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
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
  log.Println("binance futures tasks symbols flush...")
  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    config.LOCKS_TASKS_SYMBOLS_FLUSH,
  )
  if !mutex.Lock(50 * time.Second) {
    return
  }
  err = h.Repository.Flush()
  return
}
