package dydx

import (
  "context"
  "github.com/nats-io/nats.go"
  "gorm.io/gorm"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type TickersHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Nats       *nats.Conn
  Repository *repositories.TickersRepository
}

func NewTickersCommand() *cli.Command {
  var h TickersHandler
  return &cli.Command{
    Name:  "tickers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TickersHandler{
        Db:   common.NewDB(),
        Rdb:  common.NewRedis(),
        Ctx:  context.Background(),
        Nats: common.NewNats(),
      }
      h.Repository = &repositories.TickersRepository{
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
    },
  }
}

func (h *TickersHandler) Flush() error {
  log.Println("tickers flush...")
  return nil
}
