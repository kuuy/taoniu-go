package tradings

import (
  "context"
  "fmt"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"
  "time"

  "taoniu.local/cryptos/common"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type ScalpingHandler struct {
  Db               *gorm.DB
  Rdb              *redis.Client
  Ctx              context.Context
  Repository       *repositories.ScalpingRepository
  ParentRepository *futuresRepositories.ScalpingRepository
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
      h.Repository = &repositories.ScalpingRepository{
        Db: h.Db,
      }
      h.Repository.SymbolsRepository = &futuresRepositories.SymbolsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.AccountRepository = &futuresRepositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.OrdersRepository = &futuresRepositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.PositionRepository = &futuresRepositories.PositionsRepository{
        Db: h.Db,
      }
      h.ParentRepository = &futuresRepositories.ScalpingRepository{
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
  ids := h.Repository.ScalpingIds()
  for _, id := range ids {
    err := h.Repository.Flush(id)
    if err != nil {
      log.Println("scalping flush error", err)
    }
  }
  return nil
}

func (h *ScalpingHandler) Place() error {
  ids := h.ParentRepository.PlanIds(0)
  for _, id := range ids {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf("locks:binance:futures:tradings:scalping:place:%s", id),
    )
    if !mutex.Lock(30 * time.Second) {
      return nil
    }

    err := h.Repository.Place(id)
    if err != nil {
      log.Println("scalping place error", err)
    }

    mutex.Unlock()
  }
  return nil
}
