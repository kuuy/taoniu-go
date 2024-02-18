package tradings

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/commands/binance/spot/margin/cross/tradings/triggers"
  "taoniu.local/cryptos/common"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  marginRepositories "taoniu.local/cryptos/repositories/binance/spot/margin"
  crossRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross/tradings"
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
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.TriggersRepository{
        Db: h.Db,
      }
      h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.AccountRepository = &crossRepositories.AccountRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.MarginAccountRepository = &marginRepositories.AccountRepository{}
      h.Repository.OrdersRepository = &marginRepositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      triggers.NewGridsCommand(),
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "place",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.place(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *TriggersHandler) flush() error {
  symbols := h.Repository.Scan()
  for _, symbol := range symbols {
    err := h.Repository.Flush(symbol)
    if err != nil {
      log.Println("triggers flush error", err)
    }
  }
  return nil
}

func (h *TriggersHandler) place() error {
  symbols := h.Repository.Scan()
  for _, symbol := range symbols {
    err := h.Repository.Place(symbol)
    if err != nil {
      log.Println("triggers place error", err)
    }
  }
  return nil
}
