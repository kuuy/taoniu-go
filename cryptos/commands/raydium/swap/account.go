package swap

import (
  "context"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/repositories/raydium/swap"
)

type AccountHandler struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  AccountRepository *swap.AccountRepository
}

func NewAccountCommand() *cli.Command {
  var h AccountHandler
  return &cli.Command{
    Name:  "account",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AccountHandler{
        Db:  common.NewDB(3),
        Rdb: common.NewRedis(3),
        Ctx: context.Background(),
      }
      h.AccountRepository = &swap.AccountRepository{
        Db:  h.Db,
        Ctx: h.Ctx,
        MintsRepository: &swap.MintsRepository{
          Db:  h.Db,
          Rdb: h.Rdb,
          Ctx: h.Ctx,
        },
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.AccountRepository.Flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}
