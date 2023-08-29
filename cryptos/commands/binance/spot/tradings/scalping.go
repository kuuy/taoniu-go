package tradings

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"

  "taoniu.local/cryptos/common"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type ScalpingHandler struct {
  Db              *gorm.DB
  Rdb             *redis.Client
  Ctx             context.Context
  PlansRepository *spotRepositories.PlansRepository
  Repository      *repositories.ScalpingRepository
}

func NewScalpingCommand() *cli.Command {
  var h ScalpingHandler
  return &cli.Command{
    Name:  "scalping",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ScalpingHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.ScalpingRepository{
        Db: h.Db,
      }
      h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.AccountRepository = &spotRepositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      //h.Repository.PositionRepository = &spotRepositories.PositionsRepository{
      //  Db: h.Db,
      //}
      h.PlansRepository = &spotRepositories.PlansRepository{
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
  ids := h.PlansRepository.Ids(0)
  for _, id := range ids {
    err := h.Repository.Place(id)
    if err != nil {
      log.Println("scalping place error", err)
    }
  }
  return nil
}
