package strategies

import (
  "context"
  "log"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/strategies"
)

type DailyHandler struct {
  Db         *gorm.DB
  Repository *repositories.DailyRepository
}

func NewDailyCommand() *cli.Command {
  var h DailyHandler
  return &cli.Command{
    Name:  "daily",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = DailyHandler{
        Db: common.NewDB(),
      }
      h.Repository = &repositories.DailyRepository{
        Db:  h.Db,
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      return nil
    },
    Subcommands: []*cli.Command{
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
    },
  }
}

func (h *DailyHandler) atr() error {
  log.Println("daily atr processing...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.Atr(symbol)
  }
  return nil
}

func (h *DailyHandler) zlema() error {
  log.Println("daily zlema processing...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.Zlema(symbol)
  }
  return nil
}

func (h *DailyHandler) haZlema() error {
  log.Println("daily haZlema strategy...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.HaZlema(symbol)
  }
  return nil
}

func (h *DailyHandler) kdj() error {
  log.Println("daily zlema strategy...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.Kdj(symbol)
  }
  return nil
}

func (h *DailyHandler) bBands() error {
  log.Println("daily bbands strategy...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.BBands(symbol)
  }
  return nil
}
