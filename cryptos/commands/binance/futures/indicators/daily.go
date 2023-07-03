package indicators

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"
  "taoniu.local/cryptos/commands/binance/futures/indicators/daily"
  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type DailyHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.DailyRepository
}

func NewDailyCommand() *cli.Command {
  var h DailyHandler
  return &cli.Command{
    Name:  "daily",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = DailyHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.DailyRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      daily.NewRankingCommand(),
      {
        Name:  "pivot",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.pivot(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "atr",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.atr(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "zlema",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.zlema(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "ha_zlema",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.haZlema(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "kdj",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.kdj(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "bbands",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.bBands(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "volume-profile",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.volumeProfile(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "test",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.test(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *DailyHandler) atr() error {
  log.Println("daily atr processing...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.Atr(symbol, 14, 100)
  }
  return nil
}

func (h *DailyHandler) zlema() error {
  log.Println("daily zlema processing...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.Zlema(symbol, 14, 100)
  }
  return nil
}

func (h *DailyHandler) haZlema() error {
  log.Println("daily ha_zlema processing...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.HaZlema(symbol, 14, 100)
  }
  return nil
}

func (h *DailyHandler) kdj() error {
  log.Println("daily kdj indicator...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.Kdj(symbol, 9, 3, 100)
  }
  return nil
}

func (h *DailyHandler) bBands() error {
  log.Println("daily boll bands indicator...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.BBands(symbol, 14, 100)
  }
  return nil
}

func (h *DailyHandler) pivot() error {
  log.Println("daily pivot indicator...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.Pivot(symbol)
  }
  return nil
}

func (h *DailyHandler) volumeProfile() error {
  log.Println("daily volume profile indicator...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    err := h.Repository.VolumeProfile(symbol, 1440)
    if err != nil {
      log.Println("error", symbol, err)
    }
  }
  return nil
}

func (h *DailyHandler) test() error {
  log.Println("daily indicator test..")
  //timestamp := time.Now().Unix()
  //day, err := h.Repository.Day(timestamp)
  //if err != nil {
  //  return err
  //}
  symbol := "ARPAUSDT"
  fields := []string{"profit_target", "stop_loss_point", "risk_reward_ratio"}
  sortField := "risk_reward_ratio"
  sortType := -1
  current := 1
  pageSize := 20
  rank := h.Repository.Ranking(symbol, fields, sortField, sortType, current, pageSize)
  log.Println("rank", rank)
  return nil
}
