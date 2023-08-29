package dydx

import (
  "context"
  "errors"
  "fmt"
  "log"
  "strconv"
  models "taoniu.local/cryptos/models/dydx"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/dydx"
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
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
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
          side, _ := strconv.Atoi(c.Args().Get(1))
          if side != 1 && side != 2 {
            log.Fatal("side invalid")
            return nil
          }
          if err := h.Apply(symbol, side); err != nil {
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
        Name:  "scan",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Scan(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *ScalpingHandler) Apply(symbol string, side int) error {
  log.Println("dydx scalping apply...")

  capital := 5000.0

  data, _ := h.Rdb.HMGet(
    h.Ctx,
    fmt.Sprintf(
      "dydx:indicators:1d:%s:%s",
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
        "dydx:indicators:1m:%s:%s",
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
  stopPrice, _ := strconv.ParseFloat(data[1].(string), 64)

  var price float64
  if side == 1 {
    price = takePrice
  } else {
    price = stopPrice
  }

  expiredAt := time.Now().Add(time.Hour * 24 * 14)
  err := h.Repository.Apply(symbol, side, capital, price, expiredAt)
  if err != nil {
    return err
  }

  return nil
}

func (h *ScalpingHandler) Plans() error {
  log.Println("dydx scalping plans...")
  var plans []*models.Plan
  h.Db.Model(&plans).Select([]string{"symbol", "count(1) as times"}).Group("symbol").Having("count(1) > 10").Order("times desc").Find(&plans)
  for _, plan := range plans {
    h.Apply(plan.Symbol, plan.Side)
  }
  return nil
}

func (h *ScalpingHandler) Scan() error {
  log.Println("dydx scalping scan...")
  symbols := h.Repository.Scan()
  log.Println("symbols", symbols)
  return nil
}
