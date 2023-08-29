package tradings

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"

  "taoniu.local/cryptos/common"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/tradings"
)

type TriggersHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.TriggersRepository
}

func NewTriggersCommand() *cli.Command {
  var h TriggersHandler
  return &cli.Command{
    Name:  "triggers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TriggersHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.TriggersRepository{
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

func (h *TriggersHandler) Flush() error {
  ids := h.Repository.TriggerIds()
  for _, id := range ids {
    err := h.Repository.Flush(id)
    if err != nil {
      log.Println("triggers flush error", err)
    }
  }
  return nil
}

func (h *TriggersHandler) Place() error {
  ids := h.Repository.Ids()
  for _, id := range ids {
    err := h.Repository.Place(id)
    if err != nil {
      log.Println("triggers place error", err)
    }
  }
  return nil
}
