package tasks

import (
  "context"
  "fmt"
  "log"
  config "taoniu.local/cryptos/config/binance/futures"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type StrategiesHandler struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Repository        *repositories.StrategiesRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewStrategiesCommand() *cli.Command {
  var h StrategiesHandler
  return &cli.Command{
    Name:  "strategies",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = StrategiesHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.StrategiesRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
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

func (h *StrategiesHandler) Clean() error {
  log.Println("binance futures tasks strategies clean...")
  symbols := h.SymbolsRepository.Symbols()
  for _, symbol := range symbols {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TASKS_STRATEGIES_CLEAN, symbol),
    )
    if !mutex.Lock(5 * time.Second) {
      continue
    }
    h.Repository.Clean(symbol)
  }
  return nil
}
