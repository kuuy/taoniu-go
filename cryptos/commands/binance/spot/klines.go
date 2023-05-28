package spot

import (
  "context"
  "log"
  "strconv"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type KlinesHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.KlinesRepository
}

func NewKlinesCommand() *cli.Command {
  var h KlinesHandler
  return &cli.Command{
    Name:  "klines",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = KlinesHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.KlinesRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval is empty")
            return nil
          }
          limit, _ := strconv.Atoi(c.Args().Get(1))
          if interval == "1d" && (limit < 1 || limit > 100) {
            log.Fatal("limit not in 1~100")
            return nil
          }
          if interval == "1m" && (limit < 1 || limit > 1000) {
            log.Fatal("limit not in 1~1000")
            return nil
          }
          if err := h.flush(interval, limit); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "fix",
        Usage: "",
        Action: func(c *cli.Context) error {
          interval := c.Args().Get(0)
          if interval == "" {
            log.Fatal("interval is empty")
            return nil
          }
          limit, _ := strconv.Atoi(c.Args().Get(1))
          if interval == "1d" && (limit < 1 || limit > 100) {
            log.Fatal("limit not in 1~100")
            return nil
          }
          if interval == "1m" && (limit < 1 || limit > 1000) {
            log.Fatal("limit not in 1~1000")
            return nil
          }
          if err := h.fix(interval, limit); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "clean",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.clean(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *KlinesHandler) flush(interval string, limit int) error {
  log.Println("binance spot klines flush...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    err := h.Repository.Flush(symbol, interval, limit)
    if err != nil {
      log.Println("kline flush error", err)
    }
  }

  return nil
}

func (h *KlinesHandler) fix(interval string, limit int) error {
  log.Println("binance spot klines flush...")
  var symbols []string
  h.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
  for _, symbol := range symbols {
    if int(h.Repository.Count(symbol, interval)) > limit {
      continue
    }
    err := h.Repository.Flush(symbol, interval, limit)
    if err != nil {
      log.Println("kline flush error", err)
    }
  }

  return nil
}

func (h *KlinesHandler) clean() error {
  log.Println("binance spot klines daily clean...")
  h.Repository.Clean()
  return nil
}
