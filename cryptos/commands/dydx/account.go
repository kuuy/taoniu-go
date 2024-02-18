package dydx

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
)

type AccountHandler struct {
  Db         *gorm.DB
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
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.AccountRepository{
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
        Name:  "balance",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Balance(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *AccountHandler) Flush() error {
  log.Println("dydx account flush processing...")
  return h.Repository.Flush()
}

func (h *AccountHandler) Balance() error {
  log.Println("dydx account balance")
  balance, err := h.Repository.Balance()
  if err != nil {
    return err
  }
  log.Println("balnace", balance)
  return nil
}
