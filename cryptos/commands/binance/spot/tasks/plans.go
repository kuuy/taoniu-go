package tasks

import (
  "context"
  "errors"
  "fmt"
  "log"
  "slices"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type PlansHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  PlansRepository    *repositories.PlansRepository
  ScalpingRepository *repositories.ScalpingRepository
}

func NewPlansCommand() *cli.Command {
  var h PlansHandler
  return &cli.Command{
    Name:  "plans",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = PlansHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.PlansRepository = &repositories.PlansRepository{
        Db: h.Db,
      }
      h.PlansRepository.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      h.ScalpingRepository = &repositories.ScalpingRepository{
        Db: h.Db,
      }
      return nil
    },
    After: func(c *cli.Context) error {
      sqlDB, _ := h.Db.DB()
      sqlDB.Close()
      h.Rdb.Close()
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          interval := c.Args().Get(0)
          if !slices.Contains([]string{"1m", "15m", "4h", "1d"}, interval) {
            return errors.New("invalid interval")
          }
          if err := h.Flush(interval); err != nil {
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

func (h *PlansHandler) Flush(interval string) error {
  log.Println("binance futures tasks plans flush...")
  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf(config.LOCKS_PLANS_FLUSH, interval),
  )
  if !mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()
  return h.PlansRepository.Flush(interval)
}

func (h *PlansHandler) Clean() error {
  log.Println("binance spot tasks plans clean...")
  symbols := h.ScalpingRepository.Scan()
  for _, symbol := range symbols {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TASKS_PLANS_CLEAN, symbol),
    )
    if !mutex.Lock(5 * time.Second) {
      continue
    }
    h.PlansRepository.Clean(symbol)
  }
  return nil
}
