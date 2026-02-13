package futures

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type RiskManagerHandler struct {
  Db                    *gorm.DB
  Rdb                   *redis.Client
  Ctx                   context.Context
  RiskManagerRepository *repositories.RiskManagerRepository
}

func NewRiskManagerCommand() *cli.Command {
  var h RiskManagerHandler
  return &cli.Command{
    Name:  "risk-manager",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = RiskManagerHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.RiskManagerRepository = repositories.NewRiskManagerRepository(
        h.Db,
        h.Rdb,
        h.Ctx,
      )
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "calc",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(1)
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval can not be empty")
            return nil
          }
          if err := h.Calc(symbol, interval); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *RiskManagerHandler) Calc(symbol string, interval string) error {
  log.Println("binance futures risk manager calc...")
  result := h.RiskManagerRepository.Calc(symbol, interval)
  log.Println("result", result)
  return nil
}
