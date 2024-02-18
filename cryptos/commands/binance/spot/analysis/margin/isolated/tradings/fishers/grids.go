package fishers

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/analysis/margin/isolated/tradings/fishers"
)

type GridsHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.GridsRepository
}

func NewGridsCommand() *cli.Command {
  var h GridsHandler
  return &cli.Command{
    Name:  "grids",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = GridsHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.GridsRepository{
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
    },
  }
}

func (h *GridsHandler) Flush() error {
  log.Println("analysis spot margin isolated tradings fishers grids flush...")
  h.Repository.Flush()

  return nil
}
