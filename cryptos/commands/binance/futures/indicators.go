package futures

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/commands/binance/futures/indicators"
  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type IndicatorsHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.IndicatorsRepository
}

func NewIndicatorsCommand() *cli.Command {
  var h IndicatorsHandler
  return &cli.Command{
    Name:  "indicators",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = IndicatorsHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.IndicatorsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      indicators.NewAtrCommand(),
      indicators.NewPivotCommand(),
      indicators.NewKdjCommand(),
      indicators.NewRsiCommand(),
      indicators.NewStochRsiCommand(),
      indicators.NewZlemaCommand(),
      indicators.NewHaZlemaCommand(),
      indicators.NewBBandsCommand(),
      indicators.NewAndeanOscillatorCommand(),
      indicators.NewIchimokuCloudCommand(),
      indicators.NewSuperTrendCommand(),
      indicators.NewVolumeProfileCommand(),
      {
        Name:  "ranking",
        Usage: "",
        Action: func(c *cli.Context) error {
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval can not be empty")
            return nil
          }
          if err := h.Ranking(interval); err != nil {
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

func (h *IndicatorsHandler) Ranking(interval string) error {
  log.Println("indicators atr processing...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  fields := []string{
    "r3",
    "r2",
    "r1",
    "s1",
    "s2",
    "s3",
    "poc",
    "vah",
    "val",
    "profit_target",
    "take_profit_price",
    "stop_loss_point",
  }
  sortField := "poc"
  sortType := -1
  current := 1
  pageSize := 10
  result := h.Repository.Ranking(symbols, interval, fields, sortField, sortType, current, pageSize)
  log.Println("result", result)
  return nil
}

func (h *IndicatorsHandler) Flush(symbol string, interval string) (err error) {
  log.Println("indicators atr flush...")
  err = h.Repository.Atr.Flush(symbol, interval, 14, 100)
  return
}
