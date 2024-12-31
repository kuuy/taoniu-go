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
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type KlinesHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  Nats               *nats.Conn
  KlinesRepository   *repositories.KlinesRepository
  SymbolsRepository  *repositories.SymbolsRepository
  ScalpingRepository *repositories.ScalpingRepository
}

func NewKlinesCommand() *cli.Command {
  var h KlinesHandler
  return &cli.Command{
    Name:  "klines",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = KlinesHandler{
        Db:   common.NewDB(2),
        Rdb:  common.NewRedis(2),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
      }
      h.KlinesRepository = &repositories.KlinesRepository{
        Db:   h.Db,
        Rdb:  h.Rdb,
        Ctx:  h.Ctx,
        Nats: h.Nats,
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

func (h *KlinesHandler) Flush() error {
  log.Println("binance futures tasks klines flush...")
  symbols := h.ScalpingRepository.Scan(2)
  for _, symbol := range symbols {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TASKS_KLINES_FLUSH, symbol),
    )
    if !mutex.Lock(5 * time.Second) {
      continue
    }
    h.KlinesRepository.Flush(symbol, "1m", 0, 5)
    h.KlinesRepository.Flush(symbol, "15m", 0, 1)
    h.KlinesRepository.Flush(symbol, "4h", 0, 1)
    h.KlinesRepository.Flush(symbol, "1d", 0, 1)
  }
  return nil
}

func (h *KlinesHandler) Fix() error {
  log.Println("binance futures tasks klines fix...")
  symbols := h.ScalpingRepository.Scan(2)
  for _, symbol := range symbols {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TASKS_KLINES_FIX, symbol),
    )
    if !mutex.Lock(30 * time.Second) {
      continue
    }
    h.KlinesRepository.Fix(symbol, "1m", 1440)
    h.KlinesRepository.Fix(symbol, "15m", 672)
    h.KlinesRepository.Fix(symbol, "4h", 126)
    h.KlinesRepository.Fix(symbol, "1d", 100)
  }
  return nil
}

func (h *KlinesHandler) Clean() error {
  log.Println("binance futures tasks klines clean...")
  symbols := h.ScalpingRepository.Scan(2)
  for _, symbol := range symbols {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TASKS_KLINES_CLEAN, symbol),
    )
    if !mutex.Lock(5 * time.Second) {
      continue
    }
    h.KlinesRepository.Clean(symbol)
  }
  return nil
}
