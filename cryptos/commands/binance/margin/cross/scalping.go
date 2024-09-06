package cross

import (
  "context"
  "errors"
  "fmt"
  "github.com/shopspring/decimal"
  "log"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/margin/cross"
  spotModels "taoniu.local/cryptos/models/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/margin/cross"
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
        Name:  "init",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Init(); err != nil {
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
          side, _ := strconv.Atoi(c.Args().Get(0))
          if side != 1 && side != 2 {
            log.Fatal("side invalid")
            return nil
          }
          if err := h.Flush(side); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *ScalpingHandler) Apply(symbol string, side int) error {
  log.Println("margin cross scalping apply...")

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

func (h *ScalpingHandler) Init() error {
  log.Println("margin cross scalping init...")
  var positions []*spotModels.Position
  h.Db.Select([]string{"symbol", "entry_price", "entry_quantity"}).Where("entry_amount >= 0").Find(&positions)
  for _, position := range positions {
    amount, _ := decimal.NewFromFloat(position.EntryPrice).Mul(decimal.NewFromFloat(position.EntryQuantity)).Float64()
    log.Println("position", position.Symbol, position.EntryPrice, amount)
    h.Apply(position.Symbol, 1)
  }
  return nil
}

func (h *ScalpingHandler) Plans() error {
  log.Println("margin cross scalping plans...")
  var plans []*spotModels.Plan
  h.Db.Model(&plans).Select([]string{"symbol", "count(1) as times"}).Group("symbol").Having("count(1) > 10").Order("times desc").Find(&plans)
  for _, plan := range plans {
    h.Apply(plan.Symbol, plan.Side)
  }
  return nil
}

func (h *ScalpingHandler) Flush(side int) error {
  var scalping []*models.Scalping
  h.Db.Model(&models.Scalping{}).Where("side=? AND status=1", side).Find(&scalping)
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

    var price float64
    if side == 1 {
      price = takePrice
    } else {
      price = stopPrice
    }

    log.Println("scalping update", entity.Symbol, entity.Side, price)

    h.Db.Model(&entity).Update("price", price)
  }
  return nil
}
