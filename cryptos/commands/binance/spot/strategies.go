package spot

import (
  "context"
  "log"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type StrategiesHandler struct {
  Db                *gorm.DB
  Rdb               *redis.Client
  Ctx               context.Context
  Repository        *repositories.StrategiesRepository
  SymbolsRepository *repositories.SymbolsRepository
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
      h.Repository = &repositories.StrategiesRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "atr",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(1)
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval can not be empty")
            return nil
          }
          if err := h.Atr(symbol, interval); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "zlema",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(1)
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval can not be empty")
            return nil
          }
          if err := h.Zlema(symbol, interval); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "ha-zlema",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(1)
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval can not be empty")
            return nil
          }
          if err := h.HaZlema(symbol, interval); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "kdj",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(1)
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval can not be empty")
            return nil
          }
          if err := h.Kdj(symbol, interval); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "bbands",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(1)
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval can not be empty")
            return nil
          }
          if err := h.BBands(symbol, interval); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "ichimoku-cloud",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(1)
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval can not be empty")
            return nil
          }
          if err := h.IchimokuCloud(symbol, interval); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
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

func (h *StrategiesHandler) Atr(symbol string, interval string) error {
  log.Println("strategies atr processing...")
  var symbols []string
  if symbol == "" {
    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    h.Repository.Atr(symbol, interval)
  }
  return nil
}

func (h *StrategiesHandler) Zlema(symbol string, interval string) error {
  log.Println("strategies zlema processing...")
  var symbols []string
  if symbol == "" {
    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    h.Repository.Zlema(symbol, interval)
  }
  return nil
}

func (h *StrategiesHandler) HaZlema(symbol string, interval string) error {
  log.Println("strategies haZlema calc...")
  var symbols []string
  if symbol == "" {
    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    h.Repository.HaZlema(symbol, interval)
  }
  return nil
}

func (h *StrategiesHandler) Kdj(symbol string, interval string) error {
  log.Println("strategies zlema calc...")
  var symbols []string
  if symbol == "" {
    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    h.Repository.Kdj(symbol, interval)
  }
  return nil
}

func (h *StrategiesHandler) BBands(symbol string, interval string) error {
  log.Println("strategies bbands calc...")
  var symbols []string
  if symbol == "" {
    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    h.Repository.BBands(symbol, interval)
  }
  return nil
}

func (h *StrategiesHandler) IchimokuCloud(symbol string, interval string) error {
  log.Println("strategies ichimoku cloud calc...")
  var symbols []string
  if symbol == "" {
    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    h.Repository.IchimokuCloud(symbol, interval)
  }
  return nil
}

func (h *StrategiesHandler) Clean() error {
  log.Println("binance spot strategies clean...")
  symbols := h.SymbolsRepository.Symbols()
  for _, symbol := range symbols {
    h.Repository.Clean(symbol)
  }
  return nil
}
