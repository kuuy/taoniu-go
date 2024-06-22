package dydx

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type OrdersHandler struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Repository        *repositories.OrdersRepository
  MarketsRepository *repositories.MarketsRepository
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
      h.MarketsRepository = &repositories.MarketsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
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
    },
  }
}

func (h *OrdersHandler) Open() error {
  log.Println("dydx orders open...")
  symbol := "DOGE-USD"
  return h.Repository.Open(symbol)
}

func (h *OrdersHandler) Create() error {
  symbol := "DOGE-USD"
  side := "BUY"
  price := 0.0735
  quantity := 100.0
  positionSide := "LONG"
  orderId, err := h.Repository.Create(symbol, side, price, quantity, positionSide)
  if err != nil {
    return err
  }
  log.Println("orderId", orderId)
  //h.Repository.Test()
  return nil
}

func (h *OrdersHandler) Cancel() error {
  //symbol := "BTCUSDT"
  orderId := "5e19d29a8074753357099111b3037a32c0955374bd8abc0fc5940de69dbf3e1"
  err := h.Repository.Cancel(orderId)
  if err != nil {
    return err
  }
  log.Println("orderId", orderId)
  return nil
}

func (h *OrdersHandler) Flush() error {
  log.Println("dydx orders flush...")
  //symbol := "BTCUSDT"
  orderId := "3711916c36e458f7ee62d1727b697b78482b2bc14f780db8755a051a15cf11e"
  h.Repository.Flush(orderId)
  //orders, err := h.Rdb.SMembers(h.Ctx, "dydx:orders:flush").Result()
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
