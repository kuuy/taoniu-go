package dydx

import (
  "context"
  "github.com/go-redis/redis/v8"
  "log"
  "strconv"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type MarketsHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.MarketsRepository
}

func NewMarketsCommand() *cli.Command {
  var h MarketsHandler
  return &cli.Command{
    Name:  "markets",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = MarketsHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.MarketsRepository{
        Db:  h.Db,
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
          if err := h.Flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "price",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if symbol == "" {
            log.Fatal("symbol is empty")
            return nil
          }
          side, _ := strconv.Atoi(c.Args().Get(1))
          if side != 1 && side != 2 {
            log.Fatal("side invalid")
            return nil
          }
          if err := h.Price(symbol, side); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *MarketsHandler) Flush() error {
  log.Println("dydx markets flush processing...")
  return h.Repository.Flush()
}

func (h *MarketsHandler) Price(symbol string, side int) error {
  log.Println("dydx markets price processing...")
  price, err := h.Repository.Price(symbol, side)
  if err != nil {
    return err
  }
  log.Println("price", price)
  return nil
}
