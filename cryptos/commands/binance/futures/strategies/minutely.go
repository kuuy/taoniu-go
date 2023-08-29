package strategies

import (
  "context"
  "github.com/go-redis/redis/v8"
  "log"

  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/strategies"
)

type MinutelyHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.MinutelyRepository
}

func NewMinutelyCommand() *cli.Command {
  var h MinutelyHandler
  return &cli.Command{
    Name:  "minutely",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = MinutelyHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.MinutelyRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "atr",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if err := h.atr(symbol); err != nil {
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

func (h *MinutelyHandler) atr(symbol string) error {
  log.Println("daily atr processing...")
  var symbols []string
  if symbol == "" {
    h.Db.Model(models.Symbol{}).Select("symbol").Where("status=?", "TRADING").Find(&symbols)
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    err := h.Repository.Atr(symbol)
    if err != nil {
      log.Println("error", err.Error())
    }
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
