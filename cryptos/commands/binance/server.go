package binance

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance"
)

type ServerHandler struct {
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.ServerRepository
}

func NewServerCommand() *cli.Command {
  var h ServerHandler
  return &cli.Command{
    Name:  "server",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ServerHandler{
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.ServerRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "time",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.time(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *ServerHandler) time() error {
  log.Println("sync balance server time processing...")
  timestamp, err := h.Repository.Time()
  if err != nil {
    return err
  }
  log.Println("timestamp", timestamp)

  return nil
}
