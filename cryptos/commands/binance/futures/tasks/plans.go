package tasks

import (
  "context"
  "fmt"
  "log"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type PlansHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  Repository         *repositories.PlansRepository
  TradingsRepository *repositories.TradingsRepository
}

func NewPlansCommand() *cli.Command {
  var h PlansHandler
  return &cli.Command{
    Name:  "plans",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = PlansHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.PlansRepository{
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

func (h *PlansHandler) Clean() error {
  log.Println("binance futures tasks plans clean...")
  symbols := h.TradingsRepository.Scan()
  for _, symbol := range symbols {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TASKS_PLANS_CLEAN, symbol),
    )
    if !mutex.Lock(5 * time.Second) {
      continue
    }
    h.Repository.Clean(symbol)
  }
  return nil
}
