package tradings

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

type TriggersHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  TradingsRepository *tradingsRepositories.TriggersRepository
}

func NewTriggersCommand() *cli.Command {
  var h TriggersHandler
  return &cli.Command{
    Name:  "triggers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TriggersHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.TradingsRepository = &tradingsRepositories.TriggersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.TradingsRepository.SymbolsRepository = &repositories.SymbolsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.TradingsRepository.AccountRepository = &repositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.TradingsRepository.OrdersRepository = &repositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.TradingsRepository.PositionRepository = &repositories.PositionsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "place",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Place(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
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

func (h *TriggersHandler) Place() error {
  log.Println("futures tradings triggers place...")
  ids := h.TradingsRepository.Ids()
  for _, id := range ids {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TRADINGS_TRIGGERS_PLACE, id),
    )
    if !mutex.Lock(30 * time.Second) {
      return nil
    }
    err := h.TradingsRepository.Place(id)
    if err != nil {
      log.Println("error", err)
    }
    mutex.Unlock()
  }
  return nil
}

func (h *TriggersHandler) Flush() error {
  log.Println("futures tradings triggers flush...")
  ids := h.TradingsRepository.TriggerIds()
  for _, id := range ids {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TRADINGS_TRIGGERS_FLUSH, id),
    )
    if !mutex.Lock(30 * time.Second) {
      return nil
    }
    err := h.TradingsRepository.Flush(id)
    if err != nil {
      log.Println("error", err)
    }
    mutex.Unlock()
  }
  return nil
}
