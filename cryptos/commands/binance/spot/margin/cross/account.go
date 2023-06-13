package cross

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "taoniu.local/cryptos/common"

  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross"
)

type AccountHandler struct {
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.AccountRepository
}

func NewAccountCommand() *cli.Command {
  var h AccountHandler
  return &cli.Command{
    Name:  "account",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AccountHandler{
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.AccountRepository{
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
          if err := h.flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "balance",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.balance(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *AccountHandler) flush() error {
  log.Println("cross margin account flush processing...")
  return h.Repository.Flush()
}

func (h *AccountHandler) balance() error {
  log.Println("cross margin account flush processing...")
  balance, err := h.Repository.Balance("USDT")
  if err != nil {
    return err
  }
  log.Println("balance", balance)
  return nil
}
