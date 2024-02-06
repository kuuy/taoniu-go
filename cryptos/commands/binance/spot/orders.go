package spot

import (
  "context"
  "log"
  "strconv"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
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
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.OrdersRepository{
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
        Name:  "cancel",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Cancel(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "open",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Open(); err != nil {
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
        Name:  "sync",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if symbol == "" {
            log.Fatal("symbol is empty")
            return nil
          }
          limit, _ := strconv.Atoi(c.Args().Get(1))
          if limit < 1 {
            log.Fatal("limit less then 1")
            return nil
          }
          if err := h.Sync(symbol, limit); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *OrdersHandler) Create() error {
  log.Println("orders create...")
  symbol := "MATICUSDT"
  price := 0.99
  quantity := 10.10
  orderId, err := h.Repository.Create(symbol, "BUY", price, quantity)
  if err != nil {
    return err
  }
  log.Println("orderId", orderId)
  return nil
}

func (h *OrdersHandler) Cancel() error {
  log.Println("orders cancel...")
  symbol := "BTCUSDT"
  orderId := 3394265812
  err := h.Repository.Cancel(symbol, int64(orderId))
  if err != nil {
    return err
  }
  log.Println("orderId", orderId)
  return nil
}

func (h *OrdersHandler) Open() error {
  log.Println("spot open orders...")
  symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:spot:websocket:symbols").Result()
  for _, symbol := range symbols {
    log.Println("symbol:", symbol)
    h.Repository.Open(symbol)
  }
  return nil
}

func (h *OrdersHandler) Flush() error {
  log.Println("margin orders flush...")
  orders := h.Repository.Gets(map[string]interface{}{})
  for _, order := range orders {
    log.Println("order flush", order.Symbol, order.OrderID)
    h.Repository.Flush(order.Symbol, order.OrderID)
  }
  return nil
}

func (h *OrdersHandler) Sync(symbol string, limit int) error {
  log.Println("spot orders sync...")
  return h.Repository.Sync(symbol, 0, limit)
}
