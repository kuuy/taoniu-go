package isolated

import (
  "context"
  "log"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/margin/isolated"
)

type OrdersHandler struct {
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
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.OrdersRepository{
        Db:  common.NewDB(1),
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
        Name:  "fix",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Fix(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *OrdersHandler) Create() error {
  log.Println("margin isolated orders create...")
  symbol := "ZECUSDT"
  price := 40.5
  quantity := 0.28
  orderId, err := h.Repository.Create(symbol, "BUY", price, quantity)
  if err != nil {
    return err
  }
  log.Println("orderId", orderId)
  return nil
}

func (h *OrdersHandler) Cancel() error {
  log.Println("margin isolated orders cancel...")
  symbol := "BTCUSDT"
  orderId := int64(3394265812)
  err := h.Repository.Cancel(symbol, orderId)
  if err != nil {
    return err
  }
  return nil
}

func (h *OrdersHandler) Flush() error {
  log.Println("margin isolated orders flush...")
  orders, err := h.Rdb.SMembers(h.Ctx, "binance:margin:isolated:orders:flush").Result()
  if err != nil {
    return nil
  }
  for _, order := range orders {
    data := strings.Split(order, ",")
    symbol := data[0]
    orderId, _ := strconv.ParseInt(data[1], 10, 64)
    h.Repository.Flush(symbol, orderId)
  }
  return nil
}

func (h *OrdersHandler) Fix() error {
  log.Println("margin isolated orders fix...")
  return h.Repository.Fix(time.Now(), 20)
}
