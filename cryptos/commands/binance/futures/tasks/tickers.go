package tasks

import (
  "context"
  "fmt"
  "log"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type TickersHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  Nats               *nats.Conn
  Repository         *repositories.TickersRepository
  TradingsRepository *repositories.TradingsRepository
}

func NewTickersCommand() *cli.Command {
  var h TickersHandler
  return &cli.Command{
    Name:  "tickers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TickersHandler{
        Db:   common.NewDB(),
        Rdb:  common.NewRedis(),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
      }
      h.Repository = &repositories.TickersRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
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

func (h *TickersHandler) Flush() error {
  log.Println("binance tickers flush...")
  symbols := h.TradingsRepository.Scan()
  for _, symbol := range symbols {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf("locks:binance:futures:tickers:flush:%s", symbol),
    )
    if !mutex.Lock(5 * time.Second) {
      continue
    }
    err := h.Repository.Flush(symbol)
    if err != nil {
      log.Println("error", err.Error())
    }
  }
  return nil
}
