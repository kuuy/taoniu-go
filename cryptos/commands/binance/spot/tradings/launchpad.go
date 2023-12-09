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

type LaunchpadHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.LaunchpadRepository
}

func NewLaunchpadCommand() *cli.Command {
  var h LaunchpadHandler
  return &cli.Command{
    Name:  "launchpad",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = LaunchpadHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.LaunchpadRepository{
        Db: h.Db,
      }
      h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.AccountRepository = &spotRepositories.AccountRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.PositionRepository = &spotRepositories.PositionsRepository{}
      h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
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
      {
        Name:  "calc",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Calc(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *LaunchpadHandler) Place() error {
  log.Println("spot tradings launchpad place...")
  ids := h.Repository.Ids()
  for _, id := range ids {
    err := h.Repository.Place(id)
    if err != nil {
      log.Println("error", err)
    }
  }
  return nil
}

func (h *LaunchpadHandler) Flush() error {
  log.Println("spot tradings launchpad flush...")
  ids := h.Repository.LaunchpadIds()
  for _, id := range ids {
    err := h.Repository.Flush(id)
    if err != nil {
      log.Println("error", err)
    }
  }
  return nil
}

func (h *LaunchpadHandler) Calc() error {
  log.Println("spot tradings launchpad calc...")
  symbol := "STGUSDT"
  capital := 30000.0 * 100
  corePrice := 0.5149

  entity, err := h.Repository.SymbolsRepository.Get(symbol)
  if err != nil {
    return err
  }

  tickSize, stepSize, _, err := h.Repository.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return nil
  }

  buys := h.Repository.Buys(capital, corePrice, tickSize, stepSize)
  sells := h.Repository.Sells(capital, corePrice, tickSize, stepSize)
  for i := 0; i < len(buys); i++ {
    log.Println("buy", buys[i].BuyPrice, buys[i].BuyQuantity, sells[i].SellPrice, sells[i].SellQuantity)
  }
  return nil
}
