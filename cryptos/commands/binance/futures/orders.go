package futures

import (
  "context"
  "log"
  "strconv"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type OrdersHandler struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Repository        *repositories.OrdersRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewOrdersCommand() *cli.Command {
  var h OrdersHandler
  return &cli.Command{
    Name:  "orders",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = OrdersHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.OrdersRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "open",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if symbol == "" {
            log.Fatal("symbol is empty")
            return nil
          }
          if err := h.Open(symbol); err != nil {
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

func (h *OrdersHandler) Open(symbol string) error {
  log.Println("futures orders open...")
  return h.Repository.Open(symbol)
}

func (h *OrdersHandler) Create() error {
  symbol := "BTCUSDT"
  positionSide := "LONG"
  side := "BUY"
  price := 95875.3
  quantity := 0.003
  orderId, err := h.Repository.Create(symbol, positionSide, side, price, quantity)
  if err != nil {
    return err
  }
  log.Println("orderId", orderId)
  return nil
}

func (h *OrdersHandler) Cancel() error {
  symbol := "BTCUSDT"
  orderId := int64(3394265812)
  err := h.Repository.Cancel(symbol, orderId)
  if err != nil {
    return err
  }
  log.Println("orderId", orderId)
  return nil
}

func (h *OrdersHandler) Flush() error {
  log.Println("futures orders flush...")
  symbol := "ROSEUSDT"
  orderId := int64(4605220145)
  h.Repository.Flush(symbol, orderId)
  //orders := h.Repository.Gets(map[string]interface{}{})
  //for _, order := range orders {
  //  log.Println("order flush", order.Symbol, order.OrderId)
  //  h.Repository.Flush(order.Symbol, order.OrderId)
  //}
  return nil
}

func (h *OrdersHandler) Sync(symbol string, limit int) error {
  log.Println("futures orders sync...")
  return h.Repository.Sync(symbol, 0, limit)
}
