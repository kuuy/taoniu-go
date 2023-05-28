package tradings

import (
  "context"
  "log"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  plansRepositories "taoniu.local/cryptos/repositories/binance/spot/plans"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type TriggersHandler struct {
  Db              *gorm.DB
  Rdb             *redis.Client
  Ctx             context.Context
  Repository      *repositories.TriggersRepository
  PlansRepository *plansRepositories.DailyRepository
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
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
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
      h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "create",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Create(); err != nil {
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

func (h *TriggersHandler) Create() error {
  log.Println("spot tradings triggers create...")
  symbol := "DOGEUSDT"
  amount := 100.00
  buyPrice := 0.03917
  sellPrice := 0.08917
  expiredAt := time.Now()
  err := h.Repository.Create(symbol, amount, buyPrice, sellPrice, expiredAt)
  if err != nil {
    return err
  }

  return nil
}

func (h *TriggersHandler) Place() error {
  log.Println("spot tradings triggers place...")
  symbols := h.Repository.Scan()
  log.Println("symbols", symbols)
  for _, symbol := range symbols {
    err := h.Repository.Place(symbol)
    if err != nil {
      log.Println("error", err)
    }
  }
  return nil
}

func (h *TriggersHandler) Flush() error {
  log.Println("spot tradings triggers flush...")
  symbols := h.Repository.Scan()
  log.Println("symbols", symbols)
  for _, symbol := range symbols {
    err := h.Repository.Flush(symbol)
    if err != nil {
      log.Println("error", err)
    }
  }
  return nil
}