package tasks

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type AccountHandler struct {
  Rdb               *redis.Client
  Ctx               context.Context
  AccountRepository *repositories.AccountRepository
}

func NewAccountCommand() *cli.Command {
  var h AccountHandler
  return &cli.Command{
    Name:  "account",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AccountHandler{
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.AccountRepository = &repositories.AccountRepository{
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

func (h *AccountHandler) Flush() error {
  log.Println("binance spot account flush processing...")
  return h.AccountRepository.Flush()
}
