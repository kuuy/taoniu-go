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
  repositories "taoniu.local/cryptos/repositories/dydx"
  tradingsRepositories "taoniu.local/cryptos/repositories/dydx/tradings"
)

type KlinesHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  Nats               *nats.Conn
  Repository         *repositories.KlinesRepository
  MarketsRepository  *repositories.MarketsRepository
  TradingsRepository *repositories.TradingsRepository
}

func NewKlinesCommand() *cli.Command {
  var h KlinesHandler
  return &cli.Command{
    Name:  "klines",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = KlinesHandler{
        Db:   common.NewDB(1),
        Rdb:  common.NewRedis(1),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
      }
      h.Repository = &repositories.KlinesRepository{
        Db:   h.Db,
        Rdb:  h.Rdb,
        Ctx:  h.Ctx,
        Nats: h.Nats,
      }
      h.MarketsRepository = &repositories.MarketsRepository{
        Db: h.Db,
      }
      h.TradingsRepository = &repositories.TradingsRepository{
        Db: h.Db,
      }
      h.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
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
      {
        Name:  "fix",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Fix(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *KlinesHandler) Flush() error {
  log.Println("dydx klines flush...")
  symbols := h.TradingsRepository.Scan()
  for _, symbol := range symbols {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf("locks:dydx:klines:flush:%s", symbol),
    )
    if !mutex.Lock(5 * time.Second) {
      continue
    }
    h.Repository.Flush(symbol, "1m", 0, 5)
    h.Repository.Flush(symbol, "15m", 0, 1)
    h.Repository.Flush(symbol, "4h", 0, 1)
    h.Repository.Flush(symbol, "1d", 0, 1)
  }
  return nil
}

func (h *KlinesHandler) Fix() error {
  log.Println("dydx klines fix...")
  symbols := h.TradingsRepository.Scan()
  for _, symbol := range symbols {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf("locks:dydx:klines:flush:%s", symbol),
    )
    if !mutex.Lock(30 * time.Second) {
      continue
    }
    h.Repository.Fix(symbol, "1m", 1440)
    h.Repository.Fix(symbol, "15m", 672)
    h.Repository.Fix(symbol, "4h", 126)
    h.Repository.Fix(symbol, "1d", 100)
  }
  return nil
}
