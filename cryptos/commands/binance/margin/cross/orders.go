package cross

import (
  "context"
  "errors"
  "log"
  "strconv"
  repositories "taoniu.local/cryptos/repositories/binance/margin/cross"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
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
        Action: func(c *cli.Context) (err error) {
          symbol := c.Args().Get(0)
          if symbol == "" {
            err = errors.New("symbol can not be empty")
            return
          }
          side := c.Args().Get(1)
          if side != "BUY" && side != "SELL" {
            err = errors.New("side is not valid")
            return
          }
          price, _ := strconv.ParseFloat(c.Args().Get(2), 16)
          if price < 0 {
            err = errors.New("price is not valid")
            return
          }
          quantity, _ := strconv.ParseFloat(c.Args().Get(3), 16)
          if price < 0 {
            err = errors.New("quantity is not valid")
            return
          }
          if err = h.Create(symbol, side, price, quantity); err != nil {
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

func (h *OrdersHandler) Create(symbol string, side string, price float64, quantity float64) error {
  log.Println("binance margin cross orders create...", symbol, price, quantity)
  orderId, err := h.Repository.Create(symbol, side, price, quantity)
  if err != nil {
    return err
  }
  log.Println("orderId", orderId)
  return nil
}

func (h *OrdersHandler) Cancel() error {
  log.Println("binance margin cross orders cancel...")
  symbol := "BCHUSDT"
  orderId := int64(3778464925)
  err := h.Repository.Cancel(symbol, orderId)
  if err != nil {
    return err
  }
  return nil
}

func (h *OrdersHandler) Flush() error {
  log.Println("binance margin cross orders flush...")
  symbol := "MASKUSDT"
  orderId := int64(1211210069)
  h.Repository.Flush(symbol, orderId)
  //orders, err := h.Rdb.SMembers(h.Ctx, "binance:margin:cross:orders:flush").Result()
  //if err != nil {
  //  return nil
  //}
  //for _, order := range orders {
  //  data := strings.Split(order, ",")
  //  symbol := data[0]
  //  orderId, _ := strconv.ParseInt(data[1], 10, 64)
  //  h.Repository.Flush(symbol, orderId)
  //}
  return nil
}

func (h *OrdersHandler) Fix() error {
  log.Println("binance margin cross orders fix...")
  return h.Repository.Fix(time.Now(), 20)
}
