package spot

import (
  "context"
  "errors"
  "fmt"
  "log"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type ScalpingHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.ScalpingRepository
}

func NewScalpingCommand() *cli.Command {
  var h ScalpingHandler
  return &cli.Command{
    Name:  "scalping",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ScalpingHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.ScalpingRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "apply",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if symbol == "" {
            log.Fatal("symbol can not be empty")
            return nil
          }
          if err := h.Apply(symbol); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "plans",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Plans(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *ScalpingHandler) Apply(symbol string) error {
  log.Println("spot scalping apply...")

  capital := 5000.0

  data, _ := h.Rdb.HMGet(
    h.Ctx,
    fmt.Sprintf(
      "binance:spot:indicators:1d:%s:%s",
      symbol,
      time.Now().Format("0102"),
    ),
    "take_profit_price",
    "stop_loss_point",
  ).Result()
  if data[0] == nil || data[1] == nil {
    data, _ = h.Rdb.HMGet(
      h.Ctx,
      fmt.Sprintf(
        "binance:spot:indicators:1m:%s:%s",
        symbol,
        time.Now().Format("0102"),
      ),
      "take_profit_price",
      "stop_loss_point",
    ).Result()
    if data[0] == nil || data[1] == nil {
      return errors.New(fmt.Sprintf("[%s] indicators empty", symbol))
    }
  }
  takePrice, _ := strconv.ParseFloat(data[0].(string), 64)
  //stopPrice, _ := strconv.ParseFloat(data[1].(string), 64)

  expiredAt := time.Now().Add(time.Hour * 24 * 14)
  err := h.Repository.Apply(symbol, capital, takePrice, expiredAt)
  if err != nil {
    return err
  }

  return nil
}

func (h *ScalpingHandler) Plans() error {
  log.Println("spot scalping plans...")
  var plans []*models.Plan
  h.Db.Model(&plans).Select([]string{"symbol", "count(1) as times"}).Group("symbol").Having("count(1) > 10").Order("times desc").Find(&plans)
  for _, plan := range plans {
    if plan.Side != 1 {
      continue
    }
    h.Apply(plan.Symbol)
  }
  return nil
}

func (h *ScalpingHandler) Flush() error {
  var scalping []*models.Scalping
  h.Db.Model(&models.Scalping{}).Where("status", 1).Find(&scalping)
  for _, entity := range scalping {
    data, _ := h.Rdb.HMGet(
      h.Ctx,
      fmt.Sprintf(
        "binance:spot:indicators:1d:%s:%s",
        entity.Symbol,
        time.Now().Format("0102"),
      ),
      "vah",
      "val",
    ).Result()
    if len(data) == 0 || data[0] == nil || data[1] == nil {
      log.Println("indicators empty", entity.Symbol)
      continue
    }

    takePrice, _ := strconv.ParseFloat(data[0].(string), 64)
    stopPrice, _ := strconv.ParseFloat(data[1].(string), 64)

    log.Println("scalping update", entity.Symbol, takePrice, stopPrice)

    h.Db.Model(&entity).Update("price", takePrice)
  }
  return nil
}
