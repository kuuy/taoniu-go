package triggers

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/cross/tradings/triggers"
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
        Db: common.NewDB(),
      }
      h.Repository = &repositories.GridsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "pending",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.pending(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *GridsHandler) pending() error {
  log.Println("spot margin cross tradings triggers grids pending...")
  //data := h.Repository.Pending()
  //log.Println(data)
  return nil
}
