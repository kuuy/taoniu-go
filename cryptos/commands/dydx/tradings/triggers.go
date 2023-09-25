package tradings

import (
  "context"
  "log"
  "math"

  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  dydxRepositories "taoniu.local/cryptos/repositories/dydx"
  repositories "taoniu.local/cryptos/repositories/dydx/tradings"
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
      h.Repository.MarketsRepository = &dydxRepositories.MarketsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.AccountRepository = &dydxRepositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.OrdersRepository = &dydxRepositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.PositionRepository = &dydxRepositories.PositionsRepository{
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
      {
        Name:  "capital",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Capital(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *TriggersHandler) Capital() error {
  capital := 55000.0
  entryPrice := 2.001463551965
  entryQuantity := 5852.0

  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

  ipart, _ := math.Modf(capital)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }

  log.Println("capital", capital)

  result, err := h.Repository.PositionRepository.Capital(capital, entryAmount, places)
  if err != nil {
    return err
  }
  log.Println("result", result)

  return nil
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
