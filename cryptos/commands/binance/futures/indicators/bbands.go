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

type BBandsHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.BBandsRepository
}

func NewBBandsCommand() *cli.Command {
  var h BBandsHandler
  return &cli.Command{
    Name:  "bbands",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = BBandsHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.BBandsRepository{}
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

func (h *BBandsHandler) Get(symbol string, interval string) (err error) {
  log.Println("indicators bbands flush...")
  b1, b2, b3, w1, w2, w3, price, timestamp, err := h.Repository.Get(symbol, interval)
  if err != nil {
    return
  }
  log.Println("result", b1, b2, b3, w1, w2, w3, price, timestamp)
  return
}

func (h *BBandsHandler) Flush(symbol string, interval string) (err error) {
  log.Println("indicators bbands flush...")
  err = h.Repository.Flush(symbol, interval, 14, 100)
  return
}
