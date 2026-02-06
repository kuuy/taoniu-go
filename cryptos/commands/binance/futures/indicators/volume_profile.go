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

type VolumeProfileHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.VolumeProfileRepository
}

func NewVolumeProfileCommand() *cli.Command {
  var h VolumeProfileHandler
  return &cli.Command{
    Name:  "volume-profile",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = VolumeProfileHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.VolumeProfileRepository{}
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

func (h *VolumeProfileHandler) Get(symbol string, interval string) (err error) {
  log.Println("indicators volume profile flush...")
  return
}

func (h *VolumeProfileHandler) Flush(symbol string, interval string) (err error) {
  log.Println("indicators volume profile flush...")
  var limit int
  switch interval {
  case "1m":
    limit = 1440
  case "15m":
    limit = 672
  case "4h":
    limit = 126
  default:
    limit = 100
  }
  err = h.Repository.Flush(symbol, interval, limit)
  return
}
