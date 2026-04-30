package indicators

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type MvrvHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.MvrvRepository
}

func NewMvrvCommand() *cli.Command {
  var h MvrvHandler
  return &cli.Command{
    Name:  "mvrv",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = MvrvHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.MvrvRepository{}
      h.Repository.BaseRepository = repositories.BaseRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "get",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(1)
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval can not be empty")
            return nil
          }
          if err := h.Get(symbol, interval); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
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

func (h *MvrvHandler) Get(symbol string, interval string) (err error) {
  log.Println("indicators mvrv get...")
  value, price, timestamp, err := h.Repository.Get(symbol, interval)
  if err != nil {
    return
  }
  log.Println("result", value, price, timestamp)
  return
}

func (h *MvrvHandler) Flush() (err error) {
  log.Println("indicators mvrv flush...")
  err = h.Repository.Flush("BTCUSDT", "1d", 400)
  return
}
