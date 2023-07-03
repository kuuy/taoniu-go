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

type MinutelyHandler struct {
  Db         *gorm.DB
  Repository *repositories.MinutelyRepository
}

func NewMinutelyCommand() *cli.Command {
  var h MinutelyHandler
  return &cli.Command{
    Name:  "minutely",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = MinutelyHandler{
        Db: common.NewDB(),
      }
      h.Repository = &repositories.MinutelyRepository{
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

func (h *MinutelyHandler) atr() error {
  log.Println("daily atr processing...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.Atr(symbol)
  }
  return nil
}

func (h *MinutelyHandler) zlema() error {
  log.Println("daily zlema processing...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.Zlema(symbol)
  }
  return nil
}

func (h *MinutelyHandler) haZlema() error {
  log.Println("daily haZlema strategy...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.HaZlema(symbol)
  }
  return nil
}

func (h *MinutelyHandler) kdj() error {
  log.Println("daily zlema strategy...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.Kdj(symbol)
  }
  return nil
}

func (h *MinutelyHandler) bBands() error {
  log.Println("daily bbands strategy...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    h.Repository.BBands(symbol)
  }
  return nil
}
