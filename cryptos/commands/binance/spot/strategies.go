package spot

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/commands/binance/spot/strategies"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type StrategiesHandler struct {
  Db                   *gorm.DB
  Rdb                  *redis.Client
  Ctx                  context.Context
  StrategiesRepository *repositories.StrategiesRepository
  SymbolsRepository    *repositories.SymbolsRepository
}

func NewStrategiesCommand() *cli.Command {
  var h StrategiesHandler
  return &cli.Command{
    Name:  "strategies",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = StrategiesHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.StrategiesRepository = &repositories.StrategiesRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.StrategiesRepository.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      strategies.NewAtrCommand(),
      strategies.NewKdjCommand(),
      strategies.NewRsiCommand(),
      strategies.NewStochRsiCommand(),
      strategies.NewZlemaCommand(),
      strategies.NewHaZlemaCommand(),
      strategies.NewBBandsCommand(),
      strategies.NewAndeanOscillatorCommand(),
      strategies.NewIchimokuCloudCommand(),
      strategies.NewSuperTrendCommand(),
      strategies.NewSmcCommand(),
      {
        Name:  "clean",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Clean(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *StrategiesHandler) Clean() error {
  log.Println("binance spot strategies clean...")
  symbols := h.SymbolsRepository.Symbols()
  for _, symbol := range symbols {
    h.StrategiesRepository.Clean(symbol)
  }
  return nil
}
