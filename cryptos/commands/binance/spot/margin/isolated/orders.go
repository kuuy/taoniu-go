package isolated

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  marginRepositories "taoniu.local/cryptos/repositories/binance/spot/margin"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
)

type OrdersHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.OrdersRepository
}

func NewOrdersCommand() *cli.Command {
  var h OrdersHandler
  return &cli.Command{
    Name:  "orders",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = OrdersHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      parentRepository := &marginRepositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.Parent = parentRepository
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "open",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.open(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "sync",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.sync(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *OrdersHandler) open() error {
  log.Println("margin isolated open orders...")
  symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:spot:websocket:symbols").Result()
  for _, symbol := range symbols {
    log.Println("symbol:", symbol)
    h.Repository.Open(symbol)
  }
  return nil
}

func (h *OrdersHandler) sync() error {
  log.Println("margin isolated sync orders...")
  symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:spot:margin:isolated:symbols").Result()
  for _, symbol := range symbols {
    h.Repository.Sync(symbol, 100)
  }
  return nil
}
