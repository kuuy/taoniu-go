package swap

import (
  "context"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/repositories/raydium/swap"
)

type MintsHandler struct {
  Db              *gorm.DB
  Rdb             *redis.Client
  Ctx             context.Context
  MintsRepository *swap.MintsRepository
}

func NewMintsCommand() *cli.Command {
  var h MintsHandler
  return &cli.Command{
    Name:  "mints",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = MintsHandler{
        Db:  common.NewDB(3),
        Rdb: common.NewRedis(3),
        Ctx: context.Background(),
      }
      h.MintsRepository = &swap.MintsRepository{
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
          if err := h.MintsRepository.Flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}
