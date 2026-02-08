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

type StochRsiHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.StochRsiRepository
}

func NewStochRsiCommand() *cli.Command {
  var h StochRsiHandler
  return &cli.Command{
    Name:  "stoch-rsi",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = StochRsiHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.StochRsiRepository{}
      h.Repository.BaseRepository = repositories.BaseRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.Repository = &indicators.StochRsiRepository{}
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

func (h *StochRsiHandler) Flush(symbol string, interval string) (err error) {
  log.Println("strategies stoch rsi flush...")
  err = h.Repository.Flush(symbol, interval)
  return
}
