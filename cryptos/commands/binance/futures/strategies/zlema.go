package strategies

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/repositories/binance/futures/indicators"
  repositories "taoniu.local/cryptos/repositories/binance/futures/strategies"
)

type ZlemaHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.ZlemaRepository
}

func NewZlemaCommand() *cli.Command {
  var h ZlemaHandler
  return &cli.Command{
    Name:  "zlema",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ZlemaHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.ZlemaRepository{}
      h.Repository.BaseRepository = repositories.BaseRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.Repository = &indicators.ZlemaRepository{}
      h.Repository.Repository.BaseRepository = indicators.BaseRepository{
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
          symbol := c.Args().Get(1)
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval can not be empty")
            return nil
          }
          if err := h.Flush(symbol, interval); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *ZlemaHandler) Flush(symbol string, interval string) (err error) {
  log.Println("strategies zlema flush...")
  err = h.Repository.Flush(symbol, interval)
  return
}
