package tasks

import (
  "context"
  "errors"
  "fmt"
  "log"
  "slices"
  "strconv"
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
          interval := c.Args().Get(0)
          if !slices.Contains([]string{"1m", "15m", "4h", "1d"}, interval) {
            log.Fatal("interval not valid")
            return nil
          }
          current, _ := strconv.Atoi(c.Args().Get(1))
          if current < 1 {
            log.Fatal("current is less than 1")
            return nil
          }
          if err := h.Fix(interval, current); err != nil {
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

func (h *KlinesHandler) Fix(interval string, current int) (err error) {
  log.Println("binance futures tasks klines fix...", interval, current)
  symbols := h.ScalpingRepository.Scan(2)

  pageSize := common.GetEnvInt("BINANCE_SPOT_SYMBOLS_SIZE")
  startPos := (current - 1) * pageSize
  if startPos >= len(symbols) {
    err = errors.New("symbols out of range")
    return
  }
  endPos := startPos + pageSize
  if endPos > len(symbols) {
    endPos = len(symbols)
  }

  var limit int
  if interval == "1m" {
    limit = 1440
  } else if interval == "15m" {
    limit = 672
  } else if interval == "4h" {
    limit = 126
  } else if interval == "1d" {
    limit = 100
  }

  for _, symbol := range symbols[startPos:endPos] {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TASKS_KLINES_FIX, symbol),
    )
    if !mutex.Lock(30 * time.Second) {
      continue
    }
    h.KlinesRepository.Fix(symbol, interval, limit)
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
