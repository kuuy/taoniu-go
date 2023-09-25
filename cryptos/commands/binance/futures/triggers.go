package futures

import (
  "context"
  "errors"
  "fmt"
  "log"
  "strconv"
  models "taoniu.local/cryptos/models/binance/futures"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type TriggersHandler struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.TriggersRepository
}

func NewTriggersCommand() *cli.Command {
  var h TriggersHandler
  return &cli.Command{
    Name:  "triggers",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TriggersHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.TriggersRepository{
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
        Name:  "reverse",
        Usage: "",
        Action: func(c *cli.Context) error {
          side, _ := strconv.Atoi(c.Args().Get(0))
          if side != 1 && side != 2 {
            log.Fatal("side invalid")
            return nil
          }
          if err := h.Reverse(side); err != nil {
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

func (h *TriggersHandler) Apply(symbol string, side int) error {
  log.Println("futures triggers apply...")

  var capital float64
  if side == 1 {
    capital = 5000
  } else {
    capital = 1000
  }

  data, _ := h.Rdb.HMGet(
    h.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:1d:%s:%s",
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
        "binance:futures:indicators:1m:%s:%s",
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

func (h *TriggersHandler) Reverse(side int) error {
  var triggers []*models.Trigger
  h.Db.Model(&models.Trigger{}).Where("side=? AND status=1", side).Find(&triggers)
  log.Println("triggers len", len(triggers))
  for _, trigger := range triggers {
    var capital float64
    if side == 1 {
      side = 2
      capital = 1000
    } else {
      side = 1
      capital = 5000
    }

    data, _ := h.Rdb.HMGet(
      h.Ctx,
      fmt.Sprintf(
        "binance:futures:indicators:4h:%s:%s",
        trigger.Symbol,
        time.Now().Format("0102"),
      ),
      "take_profit_price",
      "stop_loss_point",
    ).Result()
    if data[0] == nil || data[1] == nil {
      log.Println("indicators empty", trigger.Symbol)
      continue
    }

    takePrice, _ := strconv.ParseFloat(data[0].(string), 64)
    stopPrice, _ := strconv.ParseFloat(data[1].(string), 64)

    log.Println("price", trigger.Symbol, takePrice, stopPrice)

    var price float64
    if side == 1 {
      price = takePrice
    } else {
      price = stopPrice
    }

    err := h.Repository.Apply(trigger.Symbol, side, capital, price, trigger.ExpiredAt)
    if err != nil {
      log.Println("reverse error", err.Error())
    }
  }

  return nil
}

func (h *TriggersHandler) Flush(side int) error {
  var triggers []*models.Trigger
  h.Db.Model(&models.Trigger{}).Where("side=? AND status=1", side).Find(&triggers)
  for _, trigger := range triggers {
    data, _ := h.Rdb.HMGet(
      h.Ctx,
      fmt.Sprintf(
        "binance:futures:indicators:1d:%s:%s",
        trigger.Symbol,
        time.Now().Format("0102"),
      ),
      "take_profit_price",
      "stop_loss_point",
    ).Result()
    if len(data) == 0 || data[0] == nil || data[1] == nil {
      log.Println("indicators empty", trigger.Symbol)
      continue
    }

    takePrice, _ := strconv.ParseFloat(data[0].(string), 64)
    stopPrice, _ := strconv.ParseFloat(data[1].(string), 64)

    var price float64
    if side == 1 {
      price = stopPrice
    } else {
      price = takePrice
    }

    log.Println("trigger update", trigger.Symbol, trigger.Side, price)

    h.Db.Model(&trigger).Update("price", price)
  }
  return nil
}
