package dydx

import (
  "context"
  "github.com/go-redis/redis/v8"
  "log"

  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type OrderbookHandler struct {
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.OrderbookRepository
}

func NewOrderbookCommand() *cli.Command {
  var h OrderbookHandler
  return &cli.Command{
    Name:  "orderbook",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = OrderbookHandler{
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.OrderbookRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if symbol == "" {
            log.Fatal("symbol is empty")
            return nil
          }
          if err := h.Flush(symbol); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *OrderbookHandler) Flush(symbol string) error {
  log.Println("dydx orderbook flush processing...")
  return h.Repository.Flush(symbol)
}
