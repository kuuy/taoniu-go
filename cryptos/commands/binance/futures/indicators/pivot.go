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

type PivotHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.PivotRepository
}

func NewPivotCommand() *cli.Command {
  var h PivotHandler
  return &cli.Command{
    Name:  "pivot",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = PivotHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.PivotRepository{}
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

func (h *PivotHandler) Get(symbol string, interval string) (err error) {
  log.Println("indicators pivot get...")
  r3, r2, r1, s1, s2, s3, err := h.Repository.Get(symbol, interval)
  if err != nil {
    return
  }
  log.Println("result", r3, r2, r1, s1, s2, s3)
  return
}

func (h *PivotHandler) Flush(symbol string, interval string) (err error) {
  log.Println("indicators pivot flush...")
  err = h.Repository.Flush(symbol, interval)
  return
}
