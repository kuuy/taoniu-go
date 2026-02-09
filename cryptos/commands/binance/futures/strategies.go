package futures

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/commands/binance/futures/strategies"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
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
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
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

//func (h *StrategiesHandler) Atr(symbol string, interval string) error {
//  log.Println("strategies atr processing...")
//  var symbols []string
//  if symbol == "" {
//    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
//  } else {
//    symbols = append(symbols, symbol)
//  }
//  for _, symbol := range symbols {
//    h.StrategiesRepository.Atr(symbol, interval)
//  }
//  return nil
//}

//func (h *StrategiesHandler) Zlema(symbol string, interval string) error {
//  log.Println("strategies zlema processing...")
//  var symbols []string
//  if symbol == "" {
//    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
//  } else {
//    symbols = append(symbols, symbol)
//  }
//  for _, symbol := range symbols {
//    h.StrategiesRepository.Zlema(symbol, interval)
//  }
//  return nil
//}
//
//func (h *StrategiesHandler) HaZlema(symbol string, interval string) error {
//  log.Println("strategies haZlema calc...")
//  var symbols []string
//  if symbol == "" {
//    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
//  } else {
//    symbols = append(symbols, symbol)
//  }
//  for _, symbol := range symbols {
//    h.StrategiesRepository.HaZlema(symbol, interval)
//  }
//  return nil
//}

//func (h *StrategiesHandler) Kdj(symbol string, interval string) error {
//  log.Println("strategies zlema calc...")
//  var symbols []string
//  if symbol == "" {
//    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
//  } else {
//    symbols = append(symbols, symbol)
//  }
//  for _, symbol := range symbols {
//    h.StrategiesRepository.Kdj(symbol, interval)
//  }
//  return nil
//}

//func (h *StrategiesHandler) BBands(symbol string, interval string) error {
//  log.Println("strategies bbands calc...")
//  var symbols []string
//  if symbol == "" {
//    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
//  } else {
//    symbols = append(symbols, symbol)
//  }
//  for _, symbol := range symbols {
//    h.StrategiesRepository.BBands(symbol, interval)
//  }
//  return nil
//}

//func (h *StrategiesHandler) IchimokuCloud(symbol string, interval string) error {
//  log.Println("strategies ichimoku cloud calc...")
//  var symbols []string
//  if symbol == "" {
//    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
//  } else {
//    symbols = append(symbols, symbol)
//  }
//  for _, symbol := range symbols {
//    h.StrategiesRepository.IchimokuCloud(symbol, interval)
//  }
//  return nil
//}

func (h *StrategiesHandler) Clean() error {
  log.Println("binance futures strategies clean...")
  symbols := h.SymbolsRepository.Symbols()
  for _, symbol := range symbols {
    h.StrategiesRepository.Clean(symbol)
  }
  return nil
}
