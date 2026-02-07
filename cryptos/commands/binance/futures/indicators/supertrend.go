package indicators

import (
  "context"
  "log"

  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  indicatorsRepositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type SuperTrendHandler struct {
  Repository *repositories.IndicatorsRepository
}

func NewSuperTrendCommand() *cli.Command {
  var h SuperTrendHandler
  return &cli.Command{
    Name:  "supertrend",
    Usage: "",
    Before: func(c *cli.Context) error {
      db := common.NewDB(2)
      rdb := common.NewRedis(2)
      ctx := context.Background()
      h.Repository = &repositories.IndicatorsRepository{
        Db:  db,
        Rdb: rdb,
        Ctx: ctx,
      }
      baseRepository := indicatorsRepositories.BaseRepository{
        Db:  db,
        Rdb: rdb,
        Ctx: ctx,
      }
      h.Repository.SuperTrend = &indicatorsRepositories.SuperTrendRepository{BaseRepository: baseRepository}
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

func (h *SuperTrendHandler) Get(symbol string, interval string) (err error) {
  log.Println("indicators supertrend get...")
  signal, superTrend, price, timestamp, err := h.Repository.SuperTrend.Get(symbol, interval)
  if err != nil {
    return
  }
  log.Println("result", signal, superTrend, price, timestamp)
  return
}

func (h *SuperTrendHandler) Flush(symbol string, interval string) (err error) {
  log.Println("indicators supertrend flush...")
  err = h.Repository.SuperTrend.Flush(symbol, interval, 10, 3.0, 100)
  return
}
