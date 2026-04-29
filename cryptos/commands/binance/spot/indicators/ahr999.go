package indicators

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/indicators"
)

type Ahr999Handler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.Ahr999Repository
}

func NewAhr999Command() *cli.Command {
  var h Ahr999Handler
  return &cli.Command{
    Name:  "ahr999",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = Ahr999Handler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.Ahr999Repository{}
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

func (h *Ahr999Handler) Get(symbol string, interval string) (err error) {
  log.Println("indicators ahr999 get...")
  value, price, timestamp, err := h.Repository.Get(symbol, interval)
  if err != nil {
    return
  }
  log.Println("result", value, price, timestamp)
  return
}

func (h *Ahr999Handler) Flush() (err error) {
  log.Println("indicators ahr999 flush...")
  err = h.Repository.Flush("BTCUSDT", "1d", 400)
  return
}
