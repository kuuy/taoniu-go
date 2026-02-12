package indicators

import (
  "context"
  "fmt"
  "log"

  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  indicatorsRepositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type SmcHandler struct {
  Repository *repositories.IndicatorsRepository
}

func NewSmcCommand() *cli.Command {
  var h SmcHandler
  return &cli.Command{
    Name:  "smc",
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
      h.Repository.Smc = &indicatorsRepositories.SmcRepository{BaseRepository: baseRepository}
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

func (h *SmcHandler) Get(symbol string, interval string) (err error) {
  log.Println("indicators smc get...")
  trend, bos, choch, high, low, obs, err := h.Repository.Smc.Get(symbol, interval)
  if err != nil {
    return
  }
  fmt.Printf("Trend: %d\nBOS: %d\nCHoCH: %d\nHigh: %f\nLow: %f\nOBs: %v\n", trend, bos, choch, high, low, obs)
  return
}

func (h *SmcHandler) Flush(symbol string, interval string) (err error) {
  log.Println("indicators smc flush...")
  err = h.Repository.Smc.Flush(symbol, interval, 5, 100)
  return
}
