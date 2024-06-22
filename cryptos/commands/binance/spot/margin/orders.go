package margin

import (
  "context"
  "log"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin"
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
          if err := h.create(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "cancel",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.cancel(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
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
        Name:  "fix",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.fix(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *OrdersHandler) create() error {
  log.Println("margin orders create...")
  symbol := "ZECUSDT"
  price := 40.5
  quantity := 0.28
  orderId, err := h.Repository.Create(symbol, "BUY", price, quantity, true)
  if err != nil {
    return err
  }
  log.Println("orderId", orderId)
  return nil
}

func (h *OrdersHandler) cancel() error {
  log.Println("margin orders cancel...")
  id := "cfi8chgv5lfbdlubgmg0"
  err := h.Repository.Cancel(id)
  if err != nil {
    return err
  }
  return nil
}

func (h *OrdersHandler) flush() error {
  log.Println("margin orders flush...")
  orders, err := h.Rdb.SMembers(h.Ctx, "binance:spot:margin:orders:flush").Result()
  if err != nil {
    return nil
  }
  for _, order := range orders {
    data := strings.Split(order, ",")
    symbol := data[0]
    orderId, _ := strconv.ParseInt(data[1], 10, 64)
    isIsolated, _ := strconv.ParseBool(data[2])
    h.Repository.Flush(symbol, orderId, isIsolated)
  }
  return nil
}

func (h *OrdersHandler) fix() error {
  log.Println("margin orders fix...")
  return h.Repository.Fix(time.Now(), 20)
}
