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

type ScalpingHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  TradingsRepository *tradingsRepositories.ScalpingRepository
  ScalpingRepository *repositories.ScalpingRepository
}

func NewScalpingCommand() *cli.Command {
  var h ScalpingHandler
  return &cli.Command{
    Name:  "scalping",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ScalpingHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.TradingsRepository = &tradingsRepositories.ScalpingRepository{
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
        Name:  "place",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Place(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *ScalpingHandler) Flush() error {
  ids := h.TradingsRepository.ScalpingIds()
  for _, id := range ids {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TRADINGS_SCALPING_FLUSH, id),
    )
    if !mutex.Lock(30 * time.Second) {
      return nil
    }

    err := h.TradingsRepository.Flush(id)
    if err != nil {
      log.Println("scalping flush error", err)
    }

    mutex.Unlock()
  }
  return nil
}

func (h *ScalpingHandler) Place() error {
  ids := h.ScalpingRepository.PlanIds(0)
  for _, id := range ids {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf(config.LOCKS_TRADINGS_SCALPING_PLACE, id),
    )
    if !mutex.Lock(30 * time.Second) {
      return nil
    }

    err := h.TradingsRepository.Place(id)
    if err != nil {
      log.Println("scalping place error", err)
    }

    mutex.Unlock()
  }
  return nil
}
