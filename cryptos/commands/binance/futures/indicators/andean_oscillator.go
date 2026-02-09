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

type AndeanOscillatorHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.AndeanOscillatorRepository
}

func NewAndeanOscillatorCommand() *cli.Command {
  var h AndeanOscillatorHandler
  return &cli.Command{
    Name:  "andean-oscillator",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AndeanOscillatorHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.AndeanOscillatorRepository{}
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

func (h *AndeanOscillatorHandler) Get(symbol string, interval string) (err error) {
  log.Println("indicators andean oscillator get...")
  bull, bear, signal, err := h.Repository.Get(symbol, interval)
  if err != nil {
    return
  }
  log.Println("result", bull, bear, signal)
  return
}

func (h *AndeanOscillatorHandler) Flush(symbol string, interval string) (err error) {
  log.Println("indicators andean oscillator flush...")
  var limit int
  if interval == "1m" {
    limit = 1440
  } else if interval == "15m" {
    limit = 672
  } else if interval == "4h" {
    limit = 126
  } else {
    limit = 100
  }
  return h.Repository.Flush(symbol, interval, 50, 9, limit)
}
