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
  indicatorsRepositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
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
      baseRepository := indicatorsRepositories.BaseRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.Repository.Atr = &indicatorsRepositories.AtrRepository{BaseRepository: baseRepository}
      h.Repository.BBands = &indicatorsRepositories.BBandsRepository{BaseRepository: baseRepository}
      h.Repository.StochRsi = &indicatorsRepositories.StochRsiRepository{BaseRepository: baseRepository}
      h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      indicators.NewAtrCommand(),
      indicators.NewPivotCommand(),
      indicators.NewBBandsCommand(),
      indicators.NewRsiStochCommand(),
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

func (h *IndicatorsHandler) Zlema(symbol string, interval string) error {
  log.Println("indicators zlema processing...")
  var symbols []string
  if symbol == "" {
    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    err := h.Repository.Zlema(symbol, interval, 14, 100)
    if err != nil {
      log.Println("error", err.Error())
    }
  }
  return nil
}

func (h *IndicatorsHandler) HaZlema(symbol string, interval string) error {
  log.Println("indicators ha_zlema processing...")
  var symbols []string
  if symbol == "" {
    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    err := h.Repository.HaZlema(symbol, interval, 14, 100)
    if err != nil {
      log.Println("error", err.Error())
    }
  }
  return nil
}

func (h *IndicatorsHandler) Kdj(symbol string, interval string) error {
  log.Println("indicators kdj calc...")
  var symbols []string
  if symbol == "" {
    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    err := h.Repository.Kdj(symbol, interval, 9, 3, 100)
    if err != nil {
      log.Println("error", err.Error())
    }
  }
  return nil
}

func (h *IndicatorsHandler) IchimokuCloud(symbol string, interval string) error {
  log.Println("indicators ichimoku cloud calc...")
  var symbols []string
  if symbol == "" {
    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    var err error
    switch interval {
    case "1m":
      err = h.Repository.IchimokuCloud(symbol, interval, 129, 374, 748, 1440)
    case "15m":
      err = h.Repository.IchimokuCloud(symbol, interval, 60, 174, 349, 672)
    case "4h":
      err = h.Repository.IchimokuCloud(symbol, interval, 11, 32, 65, 126)
    default:
      err = h.Repository.IchimokuCloud(symbol, interval, 9, 26, 52, 100)
    }
    if err != nil {
      log.Println("error", err.Error())
    }
  }
  return nil
}

func (h *IndicatorsHandler) VolumeProfile(symbol string, interval string) (err error) {
  log.Println("indicators volume profile calc...")
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
  err = h.Repository.VolumeProfile.Flush(symbol, interval, limit)
  return
}

func (h *IndicatorsHandler) Flush(symbol string, interval string) (err error) {
  log.Println("indicators atr flush...")
  err = h.Repository.Atr.Flush(symbol, interval, 14, 100)
  return
}
